export type FilterOption<T extends string> = {
  label: string;
  value: T;
};

export const deviceTypeOptions = [
  { label: "All devices", value: "all" },
  { label: "Desktop", value: "desktop" },
  { label: "Mobile", value: "mobile" },
  { label: "Tablet", value: "tablet" }
] as const satisfies ReadonlyArray<FilterOption<string>>;

export const channelOptions = [
  { label: "All channels", value: "all" },
  { label: "Web", value: "web" },
  { label: "Mobile web", value: "mobile_web" },
  { label: "App", value: "app" }
] as const satisfies ReadonlyArray<FilterOption<string>>;

export const sourceOptions = [
  { label: "All sources", value: "all" },
  { label: "Direct", value: "direct" },
  { label: "Search", value: "search" },
  { label: "Paid", value: "paid" },
  { label: "Social", value: "social" },
  { label: "Email", value: "email" }
] as const satisfies ReadonlyArray<FilterOption<string>>;

export const userTypeOptions = [
  { label: "All users", value: "all" },
  { label: "Anonymous", value: "anonymous" },
  { label: "Logged in", value: "logged_in" }
] as const satisfies ReadonlyArray<FilterOption<string>>;

export type DeviceType = (typeof deviceTypeOptions)[number]["value"];
export type Channel = (typeof channelOptions)[number]["value"];
export type TrafficSource = (typeof sourceOptions)[number]["value"];
export type UserType = (typeof userTypeOptions)[number]["value"];

export type SegmentFilters = {
  deviceType: DeviceType;
  channel: Channel;
  source: TrafficSource;
  userType: UserType;
};

export type LiveMetricsResponse = {
  activeUsersNow: number;
  sessionsToday: number;
  conversionRate: number;
  averageSessionDurationSeconds: number;
  bounceRate: number;
  productViewToCartRate: number;
};

export const defaultSegmentFilters: SegmentFilters = {
  channel: "all",
  deviceType: "all",
  source: "all",
  userType: "all"
};
