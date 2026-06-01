export type UserProfile = {
  avatar_url?: string | undefined;
  email?: string | undefined;
  full_name?: string | undefined;
  phone?: string | undefined;
  roles?: string[] | undefined;
  status?: string | undefined;
  user_id?: string | undefined;
};

export type UpdateUserProfileInput = {
  avatar_url?: string | undefined;
  full_name?: string | undefined;
  phone?: string | undefined;
};

export type NotificationPreference = {
  email_enabled?: boolean | undefined;
  marketing_enabled?: boolean | undefined;
  push_enabled?: boolean | undefined;
  sms_enabled?: boolean | undefined;
};
