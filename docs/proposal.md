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