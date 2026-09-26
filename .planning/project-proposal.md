**CU SeatFinder**  
**Project Description & Architecture Decision Records**

## **Group Members**

•   	Chatrin Yoonchalard 6631304721  
•   	Chayaphon Kultanon 6631306021  
•   	Nantaporn Phuaphanbun 6631330021

## **Problem Description**

University students lack real-time visibility into study space availability across campus libraries, forcing them to manually search between locations, which wastes valuable time and increases daily frustration. CU SeatFinder addresses this by aggregating live occupancy data from library check-in systems and giving students and staff a single place to view density, reserve meeting rooms, filter for available space, and share reviews of library conditions.

## **Target Customers**

•   	Admin: Manages library data, monitors occupancy trends, and moderates reviews.  
•   	Customer: University Student and University Staff views occupancy, reserves meeting rooms, filters libraries, and writes reviews.

## **Scenario Overview (Use Cases)**

•   	View Seat Density see how crowded each library currently is.  
•   	Reserve Meeting Room book a meeting room in advance.  
•   	Filter Empty Libraries filter to find which libraries currently have free space.  
•   	Review Library: leaves a review of a library.

 

## **Use Case Details**

### **UC-01: View Seat Density**

| Use Case Name: View Seat Density | ID: UC-01 |
| :---- | :---- |
| Importance Level: High | Use Case Type: Informational / Real-time |
| Primary Actor: Customer (Student / Staff) |  |
| Stakeholders and Interests: Customer: Wants to see live capacity percentages across libraries at a glance to choose a low-density study spot before arriving. Admin: Wants automated, accurate tracking of library traffic without manual counting. |  |
| Brief Description: The user views live occupancy percentages displayed on the library list. |  |
| Trigger: Customer opens the app home screen or navigates to the "Libraries" tab. |  |
| Type: External |  |
| Relationships: Depends on Library Turnstile Check-in/Check-out Database Service. |  |
| Precondition: Customer is on the main app screen; Library Turnstile Check-in/Check-out system is operational. |  |
| Postcondition: Customer gains clear visibility into current library capacity percentages. |  |
| Normal Flow of Events: 1\. System queries the library turnstile check-in/check-out database for current tap-in and tap-out logs. 2\. System calculates live occupancy percentages for all available libraries. 3\. System displays the main library list showing each library's name alongside its live occupancy percentage badge (e.g., Chula Central Library   45% Occupied). 4\. Customer views the list and taps on a specific library. |  |
| Subflows: None. |  |
| Alternate/Exceptional Flow: Data Delay / Database Disconnection: If real-time check-in updates are delayed or connection drops, the system displays the last calculated percentages with a "Last updated \[X\] mins ago" warning indicator on the list. |  |

 

### **UC-02: Reserve Meeting Room**

| Use Case Name: Reserve Meeting Room | ID: UC-02 |
| :---- | :---- |
| Importance Level: High | Use Case Type: Business / Transactional |
| Primary Actor: Customer (Student / Staff) |  |
| Stakeholders and Interests: Customer: Wants to secure a specific meeting room for a planned session. Admin: Wants fair usage, reduced no-shows, and optimal resource allocation. |  |
| Brief Description: The customer selects an available meeting room, sets a time duration, and confirms a meeting room reservation. |  |
| Trigger: Customer selects an available meeting room and taps "Book Meeting Room". |  |
| Type: External |  |
| Relationships: Depends on Library Database Service. |  |
| Precondition: Customer is authenticated/logged in; target meeting room is in "Available" status. |  |
| Postcondition: Meeting room status changes to "Reserved"; booking confirmation ticket/QR code is issued with a check-in countdown timer. |  |
| Normal Flow of Events: 1\. Customer selects an available meeting room on the map. 2\. Customer inputs reservation details (start time, duration). 3\. System verifies booking eligibility against quota and daily time limits. 4\. System prompts customer to confirm booking details. 5\. Customer confirms reservation. 6\. Admin accepts the reservation. 7\. System locks meeting room status to "Reserved" and generates a booking reference/QR code. |  |
| Subflows: Quota Check: System verifies if the customer has existing active bookings or active penalties. |  |
| Alternate/Exceptional Flow: Booking Conflict: If another user books the meeting room simultaneously, the system prompts the customer to select a different meeting room. Limit Reached: If the customer exceeds daily booking hours or active booking quotas, the system blocks the transaction and displays a limit error message. |  |

 

### **UC-03: Filter Empty Libraries**

| Use Case Name: Filter Empty Libraries | ID: UC-03 |
| :---- | :---- |
| Importance Level: Medium | Use Case Type: Informational / Search |
| Primary Actor: Customer (Student / Staff) |  |
| Stakeholders and Interests: Customer: Wants to quickly discover which campus libraries currently have the highest number of free seats or specific amenities. Admin: Wants to offload crowded main libraries by directing students to quieter, lesser-known locations. |  |
| Brief Description: The customer sets criteria (e.g., maximum occupancy %, minimun occupancy %) to find libraries across campus that currently have low traffic. |  |
| Trigger: Customer applies filter criteria on the main discovery/search screen. |  |
| Type: External |  |
| Relationships: Depends on Uses View Seat Density (UC-01). |  |
| Precondition: Customer is on the main discovery page. |  |
| Postcondition: System displays a filtered list of libraries ranked by availability or match percentage. |  |
| Normal Flow of Events: 1\. Customer opens search/filter controls. 2\. Customer selects filter parameters (e.g., Occupancy \< 50%). 3\. Customer submits filter query. 4\. System filters and sorts campus libraries matching criteria from least crowded to most crowded. 5\. System presents the filtered list to the customer. |  |
| Subflows: None. |  |
| Alternate/Exceptional Flow: No Match: If no library meets the criteria (e.g., exam week high traffic), the system suggests the least busy alternative location currently available. |  |

 

### **UC-04: Review Library**

| Use Case Name: Review Library | ID: UC-04 |
| :---- | :---- |
| Importance Level: Medium | Use Case Type: User-Generated Content / Transactional |
| Primary Actor: Customer (Student / Staff) |  |
| Stakeholders and Interests: Customer: Wants to share feedback on noise level, cleanliness, temperature, or wifi quality, and read peer recommendations. Admin: Wants feedback on library conditions to maintain facilities and improve service quality. |  |
| Brief Description: Customer submits a star rating, category tags (e.g., Quiet, Cold, Strong Wi-Fi), and optional written feedback for a library location. |  |
| Trigger: Customer taps "Write a Review" on a library detail page. |  |
| Type: External |  |
| Relationships: None. |  |
| Precondition: Customer is authenticated/logged in. |  |
| Postcondition: Review is posted, library average rating is recalculated, and review is flagged for admin moderation if needed. |  |
| Normal Flow of Events: 1\. Customer navigates to a library's profile page and selects "Write Review". 2\. Customer selects star rating and specific tags (e.g., Noise Level, Wi-Fi speed). 3\. Customer enters text comments and optional photos. 4\. Customer submits the review. 5\. System saves the review, updates average score display, and updates customer profile activity. |  |
| Subflows: None. |  |
| Alternate/Exceptional Flow: Inappropriate Content Filter: If text triggers automated profanity/policy filters, the system prompts the customer to revise text before submission. |  |

 

## **Functional Requirements**

### **User Registration & Login System**

•   	The system shall allow students and staff to register/log in using their university account (SSO) or student/staff ID.  
•   	The system shall allow admins to log in with elevated permissions to manage library data.  
•   	The system shall allow customers and admins to log out securely.

### **Seat Density & Occupancy System**

•   	The system shall display real-time occupancy percentages for each library, floor, or zone.  
•   	The system shall calculate occupancy based on turnstile check-in/check-out data.  
•   	The system shall display a "Last updated \[X\] mins ago" indicator when live data is delayed.

### **Meeting Room Reservation System**

•   	The system shall allow customers to reserve an available meeting room with a start time and duration.  
•   	The system shall enforce booking quotas and daily time limits per customer.  
•	The system shall allow admin to accept the customer’s reservation.  
•   	The system shall issue a booking reference/QR code with a check-in countdown timer upon confirmation.  
•   	The system shall prevent double-booking of the same meeting room and prompt the customer to choose another seat on conflict.

### **Filter & Discovery System**

•   	The system shall allow customers to filter libraries by occupancy percentage and amenities.  
•   	The system shall rank and sort filtered results from least crowded to most crowded.  
•   	The system shall suggest the next-best alternative location when no library matches the filter criteria.

### **Review System**

•   	The system shall allow customers to submit a star rating, category tags, and written feedback for a library.  
•   	The system shall recalculate and display each library's average rating after a new review is submitted.  
•   	The system shall filter submitted text against a profanity/policy filter before publishing.  
•   	The system shall allow admins to moderate (hide/remove) flagged or inappropriate reviews.

### **Admin System**

•   	The system shall allow admins to manage the reservation.  
•   	The system shall allow admins to manage moderation of customer reviews.

## **Non-Functional Requirements**

### **Operational**

•   	The system shall display properly on mobile devices.  
•   	The system shall back up occupancy, reservation, and review data daily.

### **Performance**

•   	The system shall reflect updated occupancy data within a few minutes of the source check-in/check-out event.  
•   	The system shall respond to reservation actions within 1 second.  
•   	The system shall operate with an uptime of at least 99%.

### **Security**

•   	The system shall enforce authentication via university account (SSO) with JWT-based sessions.  
•   	The system shall provide role-based access control (Admin, Customer).  
•   	The system shall protect personal and usage data in accordance with Thailand's Personal Data Protection Act (PDPA).

### **Usability**

•   	The system shall be designed with a responsive UI.  
•   	The system shall provide intuitive navigation for both customers and admins.  
•   	The system shall minimize the number of steps needed to check occupancy and reserve a meeting room.

### **Cultural**

•   	The system shall provide a Thai language user interface, with English as optional.

**Architecture Decision Records (ADRs)**

### **ADR-001: Frontend Framework**

**Context**

The occupancy list has to load fast on students' phones over campus Wi-Fi. Options: React SPA (CRA), Vue, Angular, Next.js. An SPA shows a blank screen until its JavaScript loads, and Angular is heavier than a three-person team needs. SEO is not a factor here; the whole app sits behind university SSO (ADR-004), so no page is publicly indexable. Next.js renders the first page on the server while keeping the React component model the team already knows, and its file-based routing and data-fetching conventions remove setup work the team would otherwise write by hand.

**Decision**

Next.js with server-side rendering. Live occupancy is refreshed by client-side polling, not WebSockets, matching the backend's own polling model (ADR-005).

**Status**

Accepted

**Consequences**

•  	Requires a Node runtime to deploy; static hosting is no longer an option.

•  	Needs an explicit per-route caching/revalidation policy, or Next.js's default caching will serve stale occupancy numbers.

•  	The team must learn Next.js's server vs. client rendering model.

•  	Occupancy is only ever as fresh as the last poll.

### **ADR-002: Styling Approach**

**Context**

Screens are dense (occupancy badges, filter panels, star ratings) and the team iterates on layout constantly. Options: Ant Design, Bootstrap, TailwindCSS. Ant Design's components resist customization. Bootstrap's styling looks dated. TailwindCSS applies styles as utility classes directly in markup, so layout changes need no separate stylesheet round-trip during iteration, and the app's visual identity is defined by the team rather than inherited from a library's defaults.

**Decision**

TailwindCSS utility classes, with no component library.

**Status**

Accepted

**Consequences**

•  	Every component (occupancy badges, room cards, rating widgets) must be built from scratch.

•  	Accessibility (focus states, ARIA patterns) is manual work; no component library supplies it.

•  	Long utility class lists need agreed conventions to stay readable.

•  	Visual consistency depends on team discipline rather than an enforced design system.

### **ADR-003: Backend Language & Framework**

**Context**

The API polls turnstile sources on a schedule and enforces reservation rules. Options: Node/NestJS, Django, Go. Polling a few endpoints every few minutes is not a high-concurrency workload, so throughput is not the deciding factor.  Go is chosen for static typing over the reservation logic and single-binary deployment.

**Decision**

Go, exposing a REST API.

**Status**

Accepted

**Consequences**

•  	Deployment is a single compiled binary, with no language runtime to install on the server.

•  	Smaller package ecosystem than Node; some utilities will be written in-house.

•	No team member has written Go before, so the schedule has to absorb ramp-up on its idioms and concurrency patterns.

 

### **ADR-004: Login & Identity**

**Context**

Every student and staff member already holds a university account. Options: build our own email/password signup, phone OTP, or integrate CU's SSO. Our own credential store adds signup friction and one more place passwords can leak; phone OTP carries per-message cost and still doesn't verify that the user is affiliated with the university. CU SSO verifies student/staff status as a side effect of login and keeps passwords out of our system entirely. Admin accounts, however, may belong to library staff outside the SSO directory.

**Decision**

Authenticate students and staff through CU SSO (OAuth2/OIDC), issuing JWT sessions after login, with a separate email/password path for administrative accounts. During development, run against a local OIDC-compatible mock identity provider so that login-dependent features can be built and tested without network access or real student accounts.

**Status**

Accepted

**Consequences**

•  	Auth has to sit behind an interface from day one so the mock and the real IdP are interchangeable; skipping this makes the swap a refactor.

•  	Login availability is now tied to the university SSO service's uptime, which the team does not control.

•	A second, independent auth path has to be built and secured for admin accounts outside the SSO directory.

•	No student password is ever stored, since credential handling is delegated to the IdP.

### **ADR-005: Turnstile Data Ingestion**

**Context**

No library has granted access to its turnstile/check-in data yet. Options: each library pushes updates to us (webhooks), or our backend pulls counts on a timer. Push requires every library's IT team to build something new before we can launch.

**Decision**

Poll each library's turnstile data source on a short interval, assuming an accessible export or endpoint can be negotiated.

**Status**

Proposed. Access has not been arranged, so this cannot be finalized yet.

**Consequences**

•  	Access must still be requested from each library's IT/facilities staff: unstarted work, not a formality.

•  	Needs a scheduler on the backend (Go's time.Ticker or a cron library) to drive polling.

•  	Missed tap-outs make counts drift upward over time; a reconciliation strategy is required (e.g., daily reset at closing, capping and flagging anomalies).

•  	Occupancy lags by up to one poll interval.

•  	If access turns out to be a daily CSV rather than a queryable database, this design has to be revisited.

 

### **ADR-006: Database Selection**

**Context**

The system stores three kinds of data: reservations, which need transactional writes so two customers cannot be granted the same room and time slot; reviews, which are relational; and occupancy logs from turnstile polling (ADR-005), which are append-heavy time-series rows read for recent trends.

Options: MongoDB, MySQL, PostgreSQL. MongoDB has no cross-document constraint that prevents a duplicate booking, so the guarantee would have to be enforced in application code \- the wrong place for it. MySQL and PostgreSQL both provide transactional writes; PostgreSQL was preferred for richer indexing on timestamp ranges and native JSON columns for the variable tag payloads on reviews.

For hosting, options were Prisma Postgres, Supabase, and Neon (all managed PostgreSQL). Prisma Postgres was chosen for its free tier and minimal setup, since none of these require the team to operate a database server. One consequence of that choice is worth stating: Prisma ORM, the tool usually paired with Prisma Postgres, is Node/TypeScript-only and has no supported Go client, so the Go backend (ADR-003) cannot use it.

**Decision**

PostgreSQL hosted on Prisma Postgres, queried from Go with a Go-native library (pgx or GORM) rather than Prisma ORM.

**Status**

Accepted

**Consequences**

•  	Row-level locking inside a transaction is what actually enforces the no-double-booking rule for reservations in UC-02.

•  	No Prisma ORM types or migration CLI on the Go side; migrations need a separate tool (golang-migrate, or Prisma's CLI as a dev-only step).

•  	Dependent on Prisma Postgres's managed hosting availability and free-tier limits.

•  	Occupancy logs share the relational database; heavy growth would call for extra indexing or partitioning.

 

### **ADR-007: Inter-Service Communication**

**Context**

Clients and services talk to each other in two different situations. The browser talks to the Next.js server (ADR-001), where broad tooling support and easy debugging matter. Between backend services, a handful of synchronous lookups occur: Reservation, Review, and Occupancy call Library Service to validate a room or library, and Reservation and Review call User Service `viewProfile()`. Options: REST/JSON everywhere; gRPC everywhere; or REST at the edge with gRPC between services. Browsers cannot call gRPC directly without a proxy (gRPC-Web), so gRPC at the edge would add infrastructure for no benefit. Internally, Library Service is a stable, strongly-typed master-data contract consumed by three other services and by the Next.js server, which is where a schema-first contract (protobuf) pays for itself: field changes are caught at compile time rather than at runtime. The team has not written Go before (ADR-003), so gRPC is limited to one service to keep the learning surface small. The course also requires both REST and gRPC services with CRUD.

**Decision**

REST/JSON for the browser-facing API and for the Review, Reservation, Occupancy, and User services. Library Service is gRPC-only, defined in `proto/library/v1/library.proto` with full CRUD for libraries and rooms; Occupancy, Reservation, Review, and the Next.js server call it over gRPC. The Next.js server exposes `/api/libraries` and `/api/rooms` as REST route handlers that translate to gRPC, so the browser never speaks gRPC. Review Service is the REST service with full CRUD. Service-to-service calls to User Service remain REST.

**Status**

Accepted

**Consequences**

•  	Two protocols to run and debug; gRPC calls need `grpcurl` or server reflection instead of curl.

•  	Protobuf code generation is added to the Go build; generated code is committed so teammates need no toolchain to build.

•  	The Library contract is type-checked across Go and Next.js consumers, and admin-only writes are enforced by a server interceptor reading the caller's JWT from gRPC metadata.

•  	Next.js needs a small REST-to-gRPC translation layer, which is the only place the two protocols meet.

•  	Each synchronous call adds latency and a failure point; callers set deadlines and return a clear error when a dependency is down.

•  	If more services later need typed contracts, they can adopt gRPC one at a time without changing the browser API.

### **ADR-008: Data Ownership Between Services**

**Context**

Microservices are meant to own their data so they can change independently, which argues for a separate database per service. But ADR-006 selected a single managed Prisma Postgres instance on a free tier, and running five database instances would multiply cost and operational work for a three-person team. Options: one shared schema all services read freely; a separate database instance per service; or one instance with a separate schema per service.

**Decision**

One PostgreSQL instance with one schema per service (`users`, `library`, `occupancy`, `reservation`, `review`). A service reads and writes only its own schema; data owned by another service is obtained through that service's API (ADR-007), never by querying its tables. Each service creates and migrates its own schema.

**Status**

Accepted

**Consequences**

•  	Logical data ownership, the core property of database-per-service, is kept at the cost of a single instance.

•  	No cross-service foreign keys or joins; references such as `reservation.user_id` and `review.library_id` are validated through API calls at write time and can become dangling if the referenced record is deleted.

•  	All services share one instance's availability, connection limit, and free-tier quota; a runaway query in one service can slow the others.

•  	Splitting a schema into its own instance later requires only a connection-string change for that service, because no other service touches its tables.

•  	The schema boundary is a convention, not an enforced rule, unless each service is given its own database role restricted to its schema; that hardening is deferred.
