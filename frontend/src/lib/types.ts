// Shapes mirror .planning/API-CONTRACT.md exactly (snake_case, as returned by the Go services).

export type Role = "customer" | "admin";

export type User = {
  id: string;
  student_id?: string;
  email?: string;
  name: string;
  role: Role;
  language: "th" | "en";
  created_at: string;
};

export type Zone = { id: string; name: string; capacity: number };
export type Floor = { id: string; name: string; zones: Zone[] };

export type Library = {
  id: string;
  name_th: string;
  name_en: string;
  capacity: number;
  amenities: string[];
  floors: Floor[];
  image_url?: string;
};

export type Room = {
  id: string;
  library_id: string;
  name: string;
  capacity: number;
  amenities: string[];
};

export type Occupancy = {
  library_id: string;
  current: number;
  capacity: number;
  percent: number;
  updated_at: string;
  stale: boolean;
};

export type OccupancyResponse = {
  items: Occupancy[];
  suggestion: Occupancy | null;
};

export type ReservationStatus = "pending" | "accepted" | "rejected" | "cancelled";

export type Reservation = {
  id: string;
  user_id: string;
  room_id: string;
  start_time: string;
  end_time: string;
  status: ReservationStatus;
  reference?: string;
  checkin_deadline?: string;
  created_at: string;
};

export const REVIEW_TAGS = [
  "quiet",
  "noisy",
  "cold",
  "hot",
  "strong_wifi",
  "weak_wifi",
  "clean",
  "many_outlets",
  "crowded",
] as const;
export type ReviewTag = (typeof REVIEW_TAGS)[number];

export type Review = {
  id: string;
  library_id: string;
  user_id: string;
  user_name: string;
  rating: number;
  tags: string[];
  comment?: string;
  hidden: boolean;
  created_at: string;
};

export type ReviewsResponse = { average: number; count: number; items: Review[] };

export type ApiErrorBody = { error: string; message: string };
