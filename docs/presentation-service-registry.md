# สคริปต์พรีเซนต์: Service Registry (Consul)

เวลารวม 3 นาที อ้างอิง ADR-009 และ Step 4 ใน [Proposal.md](Proposal.md)

| ช่วง | เวลา |
| :---- | :---- |
| 1. ปัญหาเดิม | 0:00 – 0:25 |
| 2. Service registry ทำงานอย่างไร | 0:25 – 1:25 |
| 3. เดโม scale และ failover | 1:25 – 2:40 |
| 4. Trade-off | 2:40 – 3:00 |

## เตรียมก่อนพรีเซนต์

```bash
cd backend && make up                       # รัน stack ทั้งหมดใน Docker
docker compose --profile app ps             # ทุกตัวต้องขึ้น (healthy)
```

เปิด Consul UI ไว้ล่วงหน้าที่ http://localhost:8500 และเตรียม terminal ไว้ที่โฟลเดอร์ของโปรเจกต์

---

## 1. ปัญหาเดิม (25 วินาที)

> "เดิมเราให้ service หากันผ่านชื่อ DNS ของ Docker ซึ่งมีปัญหาสองข้อ
> ข้อแรก DNS ไม่รู้ว่า container ไหนยังทำงานอยู่ ถ้า service ตาย request ก็ยังถูกส่งไปหา
> จนกว่า Docker จะ restart เสร็จ ข้อสอง จะเพิ่ม instance ไม่ได้ เพราะไม่มีใครรู้ว่า
> แต่ละ instance อยู่ที่ไหนและพร้อมรับงานหรือยัง"

## 2. Service registry ทำงานอย่างไร (60 วินาที)

> "เราจึงเพิ่ม Consul เป็น service registry ซึ่งทำงานสามขั้น
>
> **ขั้นที่ 1, Register:** service ทุกตัวลงทะเบียนตัวเองใน Consul ตอน start
> โดยบอกชื่อ IP port และวิธีเช็ค health service ลงทะเบียนซ้ำทุก 30 วินาที
> และถอนตัวเองตอนปิด
>
> **ขั้นที่ 2, Health check:** Consul เรียก `/healthz` ของแต่ละ instance ทุก 5 วินาที
> (ส่วน Library ที่เป็น gRPC ใช้ TCP check) instance ที่เช็คไม่ผ่านจะถูกตัดออกจากรายชื่อ
>
> **ขั้นที่ 3, Discover:** ผู้เรียกระบุ service ด้วยชื่อ เช่น `consul://library`
> แทนการเขียน host:port ตายตัว แล้วถาม Consul ว่ามี instance ไหนที่ health check ผ่านบ้าง
> จากนั้นเลือกแบบ round-robin แบบนี้เรียกว่า client-side discovery ทั้ง API Gateway
> และการเรียกกันระหว่าง service ใช้กลไกเดียวกัน ฝั่ง gRPC เราเขียน resolver
> ให้หา instance ผ่าน Consul ได้โดยตรง
>
> โค้ดอยู่ที่ `backend/internal/registry` และเรียก HTTP API ของ Consul เองโดยไม่ใช้ library เพิ่ม"

(ชี้ที่ Consul UI: มี service ครบ 6 ตัวคือ gateway, user, library, occupancy,
reservation, review และทุกตัวเป็นสีเขียว)

## 3. เดโม (75 วินาที)

### 3.1 Scale: เพิ่ม instance โดยไม่แก้ config (35 วินาที)

```bash
docker compose --profile app up -d --scale library=3
```

(รีเฟรช Consul UI: library มี 3 instance)

```bash
for i in 1 2 3 4 5 6; do curl -s -o /dev/null -w "%{http_code} " localhost:8080/api/libraries; done
docker compose logs --tail 6 gateway                # upstream= สลับกันไป 3 IP
```

> "instance ใหม่ลงทะเบียนตัวเองและได้รับ traffic ทันทีที่ health check ผ่าน
> จะเห็นว่า request กระจายไปครบทั้ง 3 ตัวโดยไม่ได้แก้ config เลย"

### 3.2 Failover: instance ตาย ระบบยังทำงาน (40 วินาที)

```bash
docker kill cu-seat-finder-library-2
for i in $(seq 1 10); do curl -s -o /dev/null -w "%{http_code} " localhost:8080/api/libraries; done
```

(ผลต้องเป็น 200 ทั้งหมด จากนั้นชี้ Consul UI ว่า instance ที่ kill เป็นสีแดง critical)

> "Consul รู้ภายในประมาณ 5 วินาทีและเลิกส่ง traffic ไปที่ตัวนั้น ระหว่างนั้นถ้าผู้เรียก
> ต่อไม่ติด จะลองใหม่กับ instance อื่นอีกหนึ่งครั้ง ผู้ใช้จึงไม่เห็น error เลย
> ข้อนี้รองรับ NFR uptime 99%"

## 4. Trade-off (20 วินาที)

> "ข้อเสียคือมี Consul เพิ่มมาหนึ่งตัวที่ต้องดูแลและอยู่บน critical path
> แต่ถ้า Consul ล่ม service จะลงทะเบียนใหม่เองภายใน 30 วินาทีหลัง Consul กลับมา
> ส่วน database กับ RabbitMQ มีตัวเดียวจึงยังใช้ DNS ของ Docker
> และถ้าย้ายไป Kubernetes ก็ใช้ Service DNS แทน Consul ได้"

---

คืนสภาพหลังเดโม (ไม่ต้องพูด):

```bash
docker compose --profile app up -d --scale library=1
```

## คำถามที่อาจโดนถาม

| คำถาม | คำตอบ |
| :---- | :---- |
| ทำไมไม่ใช้ Docker DNS ต่อ | DNS ไม่รู้ health จึงส่ง request ไปที่ container ที่ตายได้ และกระจายโหลดได้แค่ DNS round-robin |
| ทำไมเลือก Consul แทน Eureka | Consul มี health check ในตัว รันเป็น container เดียว และมี HTTP API ที่เรียกจาก Go ได้โดยไม่ต้องใช้ library ส่วน Eureka เหมาะกับ Java/Spring |
| ทำไมไม่ใช้ Kubernetes | แก้ปัญหาได้เหมือนกัน แต่เรียนรู้และดูแลยากเกินเวลาของโปรเจกต์ |
| client-side หรือ server-side discovery | client-side: ผู้เรียกถาม registry เองแล้วเลือก instance เอง |
| ถามหา Consul ทุก request ช้าไหม | cache คำตอบไว้ 1 วินาที จึงไม่ต้องถาม Consul ทุก request |
| Consul ล่มแล้วจะเป็นอย่างไร | ผู้เรียกใช้คำตอบใน cache ได้แค่ 1 วินาที พอ Consul กลับมา service จะลงทะเบียนใหม่เองภายใน 30 วินาที |
