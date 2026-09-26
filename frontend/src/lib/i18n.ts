// Tiny hand-rolled i18n: flat dictionary + {var} interpolation. No library (per ADR scope).
export type Lang = "th" | "en";
export const LANG_COOKIE = "cusf_lang";
export const DEFAULT_LANG: Lang = "th";

const dict = {
  "nav.skipToContent": { th: "ข้ามไปยังเนื้อหาหลัก", en: "Skip to main content" },
  "nav.home": { th: "ห้องสมุด", en: "Libraries" },
  "nav.reservations": { th: "การจองของฉัน", en: "My Reservations" },
  "nav.admin": { th: "แอดมิน", en: "Admin" },
  "nav.login": { th: "เข้าสู่ระบบ", en: "Log in" },
  "nav.logout": { th: "ออกจากระบบ", en: "Log out" },

  "auth.title": { th: "เข้าสู่ระบบ CU SeatFinder", en: "Log in to CU SeatFinder" },
  "auth.tab.cu": { th: "บัญชี CU", en: "CU Account" },
  "auth.tab.admin": { th: "แอดมิน", en: "Admin" },
  "auth.studentId": { th: "รหัสนิสิต/บุคลากร", en: "Student / Staff ID" },
  "auth.studentIdPlaceholder": { th: "เช่น 6631304721", en: "e.g. 6631304721" },
  "auth.name": { th: "ชื่อ-นามสกุล", en: "Full name" },
  "auth.namePlaceholder": { th: "ชื่อของคุณ", en: "Your name" },
  "auth.email": { th: "อีเมล", en: "Email" },
  "auth.password": { th: "รหัสผ่าน", en: "Password" },
  "auth.loginCu": { th: "เข้าสู่ระบบด้วยบัญชี CU", en: "Log in with CU Account" },
  "auth.loginAdmin": { th: "เข้าสู่ระบบแอดมิน", en: "Log in as Admin" },
  "auth.loggingIn": { th: "กำลังเข้าสู่ระบบ...", en: "Logging in..." },
  "auth.mockNote": {
    th: "จำลองการเข้าสู่ระบบ CU SSO ด้วยรหัสนิสิตและชื่อ",
    en: "Mock CU SSO login using your ID and name.",
  },

  "error.generic": { th: "เกิดข้อผิดพลาด กรุณาลองใหม่", en: "Something went wrong. Please try again." },
  "error.network": {
    th: "เชื่อมต่อเซิร์ฟเวอร์ไม่ได้ในขณะนี้",
    en: "Can't reach the server right now.",
  },
  "error.unauthorized": { th: "กรุณาเข้าสู่ระบบก่อน", en: "Please log in first." },
  "error.forbidden": { th: "คุณไม่มีสิทธิ์เข้าถึงหน้านี้", en: "You don't have access to this page." },
  "action.retry": { th: "ลองอีกครั้ง", en: "Retry" },
  "action.edit": { th: "แก้ไข", en: "Edit" },
  "action.delete": { th: "ลบ", en: "Delete" },
  "action.cancel": { th: "ยกเลิก", en: "Cancel" },
  "action.keep": { th: "เก็บไว้", en: "Keep" },
  "action.save": { th: "บันทึกการแก้ไข", en: "Save changes" },
  "action.saving": { th: "กำลังบันทึก...", en: "Saving..." },

  "home.title": { th: "ห้องสมุดจุฬาฯ", en: "Chula Libraries" },
  "home.minOccupancy": { th: "ความหนาแน่นขั้นต่ำ", en: "Min occupancy" },
  "home.maxOccupancy": { th: "ความหนาแน่นขั้นสูง", en: "Max occupancy" },
  "home.amenities": { th: "สิ่งอำนวยความสะดวก", en: "Amenities" },
  "home.noResults": {
    th: "ไม่พบห้องสมุดที่ตรงกับตัวกรอง",
    en: "No library matches your filters.",
  },
  "home.suggestion": { th: "ห้องสมุดที่ว่างที่สุดตอนนี้", en: "Least busy alternative right now" },
  "home.lastUpdated": { th: "อัปเดตล่าสุด {n} นาทีที่แล้ว", en: "Last updated {n} mins ago" },
  "home.freeSeats": { th: "ว่าง {n} ที่นั่ง", en: "{n} seats free" },
  "home.filtersToggle": { th: "ตัวกรอง", en: "Filters" },
  "home.filtersActive": { th: "ตัวกรอง ({n} ใช้งานอยู่)", en: "Filters ({n} active)" },
  "home.resetFilters": { th: "ล้างตัวกรอง", en: "Reset filters" },
  "home.resultCount": { th: "พบ {n} แห่ง", en: "{n} libraries" },
  "occupancy.low": { th: "ว่าง", en: "Low" },
  "occupancy.medium": { th: "ปานกลาง", en: "Moderate" },
  "occupancy.high": { th: "แน่น", en: "Busy" },
  "occupancy.full": { th: "เต็ม", en: "Full" },
  "occupancy.ariaBar": { th: "ระดับความหนาแน่น", en: "Occupancy level" },

  "library.floors": { th: "ชั้นและโซน", en: "Floors & zones" },
  "library.rooms": { th: "ห้องประชุม", en: "Meeting rooms" },
  "library.book": { th: "จอง", en: "Book" },
  "library.reviews": { th: "รีวิว", en: "Reviews" },
  "library.writeReview": { th: "เขียนรีวิว", en: "Write a review" },
  "library.noReviews": { th: "ยังไม่มีรีวิว", en: "No reviews yet." },
  "library.notFound": { th: "ไม่พบห้องสมุดนี้", en: "Library not found." },

  "review.rating": { th: "ให้คะแนน", en: "Rating" },
  "review.tags": { th: "แท็ก", en: "Tags" },
  "review.comment": { th: "ความคิดเห็น (ไม่บังคับ)", en: "Comment (optional)" },
  "review.commentPlaceholder": {
    th: "บอกเล่าประสบการณ์ของคุณ...",
    en: "Share your experience...",
  },
  "review.submit": { th: "ส่งรีวิว", en: "Submit review" },
  "review.submitting": { th: "กำลังส่ง...", en: "Submitting..." },
  "review.profanity": {
    th: "ข้อความมีคำไม่เหมาะสม กรุณาแก้ไขข้อความแล้วลองส่งอีกครั้ง",
    en: "Your comment contains inappropriate language. Please revise it and try again.",
  },
  "review.success": { th: "ขอบคุณสำหรับรีวิว!", en: "Thanks for your review!" },
  "review.loginRequired": {
    th: "กรุณาเข้าสู่ระบบเพื่อเขียนรีวิว",
    en: "Log in to write a review.",
  },
  "review.deleteConfirm": {
    th: "ลบรีวิวนี้ใช่ไหม? การลบไม่สามารถกู้คืนได้",
    en: "Delete this review? This can't be undone.",
  },
  "review.deleteYes": { th: "ลบรีวิว", en: "Delete review" },
  "tag.quiet": { th: "เงียบ", en: "Quiet" },
  "tag.noisy": { th: "เสียงดัง", en: "Noisy" },
  "tag.cold": { th: "หนาว", en: "Cold" },
  "tag.hot": { th: "ร้อน", en: "Hot" },
  "tag.strong_wifi": { th: "wifi แรง", en: "Strong wifi" },
  "tag.weak_wifi": { th: "wifi อ่อน", en: "Weak wifi" },
  "tag.clean": { th: "สะอาด", en: "Clean" },
  "tag.many_outlets": { th: "ปลั๊กเยอะ", en: "Many outlets" },
  "tag.crowded": { th: "คนเยอะ", en: "Crowded" },

  "reserve.title": { th: "จองห้องประชุม", en: "Reserve meeting room" },
  "reserve.startTime": { th: "เวลาเริ่ม", en: "Start time" },
  "reserve.duration": { th: "ระยะเวลา", en: "Duration" },
  "reserve.min30": { th: "30 นาที", en: "30 minutes" },
  "reserve.min60": { th: "1 ชั่วโมง", en: "1 hour" },
  "reserve.min90": { th: "1 ชั่วโมง 30 นาที", en: "1.5 hours" },
  "reserve.min120": { th: "2 ชั่วโมง", en: "2 hours" },
  "reserve.review": { th: "ตรวจสอบรายละเอียด", en: "Review details" },
  "reserve.confirm": { th: "ยืนยันการจอง", en: "Confirm booking" },
  "reserve.confirming": { th: "กำลังจอง...", en: "Booking..." },
  "reserve.back": { th: "ย้อนกลับ", en: "Back" },
  "reserve.success": {
    th: "ส่งคำขอจองแล้ว รอแอดมินยืนยัน",
    en: "Booking requested. Waiting for admin approval.",
  },
  "reserve.viewMyReservations": { th: "ไปที่การจองของฉัน", en: "Go to my reservations" },
  "reserve.conflict": {
    th: "ห้องนี้ถูกจองในช่วงเวลานี้แล้ว กรุณาเลือกห้องอื่นหรือเวลาอื่น",
    en: "This room is already booked for that time. Please pick another room or time.",
  },
  "reserve.quota": {
    th: "คุณมีการจองที่ยังไม่เสร็จสิ้นครบโควตาแล้ว หรือครบเวลาจองต่อวันแล้ว",
    en: "You've reached your active booking quota or daily time limit.",
  },
  "reserve.penalty": {
    th: "บัญชีของคุณถูกระงับสิทธิ์การจองชั่วคราวจากการไม่มาใช้ห้องตามนัด",
    en: "Your account is temporarily blocked from booking due to a past no-show.",
  },
  "reserve.notFound": { th: "ไม่พบห้องนี้", en: "Room not found." },

  "reservations.title": { th: "การจองของฉัน", en: "My reservations" },
  "reservations.empty": { th: "คุณยังไม่มีการจอง", en: "You have no reservations yet." },
  "reservations.status.pending": { th: "รอดำเนินการ", en: "Pending" },
  "reservations.status.accepted": { th: "ยืนยันแล้ว", en: "Accepted" },
  "reservations.status.rejected": { th: "ถูกปฏิเสธ", en: "Rejected" },
  "reservations.status.cancelled": { th: "ยกเลิกแล้ว", en: "Cancelled" },
  "reservations.checkinBy": { th: "เช็คอินภายใน", en: "Check in within" },
  "reservations.expired": { th: "หมดเวลาเช็คอินแล้ว", en: "Check-in window expired" },
  "reservations.cancel": { th: "ยกเลิกการจอง", en: "Cancel booking" },
  "reservations.reference": { th: "รหัสอ้างอิง", en: "Reference code" },
  "reservations.browse": { th: "ไปเลือกห้องสมุด →", en: "Browse libraries →" },

  "admin.title": { th: "แผงควบคุมแอดมิน", en: "Admin dashboard" },
  "admin.pending": { th: "การจองที่รอดำเนินการ", en: "Pending reservations" },
  "admin.noPending": { th: "ไม่มีการจองที่รอดำเนินการ", en: "No pending reservations." },
  "admin.accept": { th: "ยืนยัน", en: "Accept" },
  "admin.reject": { th: "ปฏิเสธ", en: "Reject" },
  "admin.room": { th: "ห้อง {id}", en: "Room {id}" },
  "admin.reviews": { th: "จัดการรีวิว", en: "Review moderation" },
  "admin.noFlagged": { th: "ไม่มีรีวิวที่ถูกรายงาน", en: "No flagged reviews." },
  "admin.hide": { th: "ซ่อน", en: "Hide" },
  "admin.unhide": { th: "เลิกซ่อน", en: "Unhide" },
  "admin.hidden": { th: "ซ่อนอยู่", en: "Hidden" },

  "adminLib.title": { th: "ห้องสมุดและห้องประชุม", en: "Libraries & rooms" },
  "adminLib.add": { th: "เพิ่มห้องสมุด", en: "Add library" },
  "adminLib.create": { th: "สร้างห้องสมุด", en: "Create library" },
  "adminLib.empty": {
    th: "ยังไม่มีห้องสมุด กด “เพิ่มห้องสมุด” เพื่อเริ่มต้น",
    en: "No libraries yet. Add one to get started.",
  },
  "adminLib.nameTh": { th: "ชื่อ (ไทย)", en: "Name (Thai)" },
  "adminLib.nameEn": { th: "ชื่อ (อังกฤษ)", en: "Name (English)" },
  "adminLib.capacity": { th: "ความจุ (ที่นั่ง)", en: "Capacity (seats)" },
  "adminLib.seats": { th: "{n} ที่นั่ง", en: "{n} seats" },
  "adminLib.addFloor": { th: "เพิ่มชั้น", en: "Add floor" },
  "adminLib.floorName": { th: "ชื่อชั้น {n}", en: "Floor {n} name" },
  "adminLib.removeFloor": { th: "ลบชั้น {n}", en: "Remove floor {n}" },
  "adminLib.addZone": { th: "เพิ่มโซน", en: "Add zone" },
  "adminLib.zoneName": { th: "ชื่อโซน {n}", en: "Zone {n} name" },
  "adminLib.zoneCapacity": { th: "ความจุโซน {n}", en: "Zone {n} capacity" },
  "adminLib.removeZone": { th: "ลบโซน {n}", en: "Remove zone {n}" },
  "adminLib.deleteConfirm": {
    th: "ลบ “{name}” ใช่ไหม? การลบไม่สามารถกู้คืนได้",
    en: "Delete “{name}”? This can't be undone.",
  },
  "adminLib.hasRooms": {
    th: "ลบไม่ได้ เพราะห้องสมุดนี้ยังมีห้องประชุมอยู่ กรุณาลบห้องทั้งหมดก่อน",
    en: "Can't delete this library while it still has rooms. Delete its rooms first.",
  },
  "adminLib.rooms": { th: "ห้องประชุม", en: "Rooms" },
  "adminLib.noRooms": { th: "ห้องสมุดนี้ยังไม่มีห้องประชุม", en: "No rooms in this library yet." },
  "adminLib.addRoom": { th: "เพิ่มห้อง", en: "Add room" },
  "adminLib.createRoom": { th: "สร้างห้อง", en: "Create room" },
  "adminLib.roomName": { th: "ชื่อห้อง", en: "Room name" },
} satisfies Record<string, Record<Lang, string>>;

export type DictKey = keyof typeof dict;

export function t(lang: Lang, key: DictKey, vars?: Record<string, string | number>): string {
  let s = dict[key]?.[lang] ?? key;
  if (vars) {
    for (const [k, v] of Object.entries(vars)) s = s.replaceAll(`{${k}}`, String(v));
  }
  return s;
}
