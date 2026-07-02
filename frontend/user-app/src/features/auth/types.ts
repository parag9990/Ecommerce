export type AuthRole = 'buyer' | 'seller';

export type OtpChannel = 'email' | 'phone';

export type OtpPurpose =
  | 'email_verify'
  | 'login'
  | 'password_reset'
  | 'phone_verify'
  | 'signup';

export type SignupRequest = {
  email: string;
  full_name: string;
  password: string;
  phone?: string;
  role?: AuthRole;
};

export type LoginRequest = {
  identifier: string;
  password: string;
  device?: {
    anonymous_id?: string;
    channel?: string;
    device_fingerprint_hash?: string;
    fingerprint?: string;
    locale?: string;
  };
};

export type TokenResponse = {
  access_token: string;
  expires_in: number;
  refresh_token: string;
};

export type UserProfile = {
  email?: string;
  full_name?: string;
  phone?: string;
  roles?: string[];
  status?: string;
  user_id?: string;
};

export type AuthSessionResponse = {
  session_id?: string;
  tokens?: TokenResponse;
  user?: UserProfile;
};

export type SendOtpRequest = {
  channel: OtpChannel;
  purpose: OtpPurpose;
  target: string;
};

export type OtpChallengeResponse = {
  challenge_id: string;
  expires_in: number;
};

export type VerifyOtpRequest = {
  challenge_id: string;
  otp: string;
};

export type ForgotPasswordRequest = {
  identifier: string;
};

export type ResetPasswordRequest = {
  challenge_id: string;
  new_password: string;
  otp: string;
};

export type SuccessResponse = {
  success: boolean;
};
