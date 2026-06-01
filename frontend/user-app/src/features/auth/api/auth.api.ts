import { postJson } from '../../../lib/http';
import type {
  AuthSessionResponse,
  ForgotPasswordRequest,
  LoginRequest,
  OtpChallengeResponse,
  ResetPasswordRequest,
  SendOtpRequest,
  SignupRequest,
  SuccessResponse,
  VerifyOtpRequest,
} from '../types';

const authEndpoints = {
  forgotPassword: '/api/v1/auth/password/forgot',
  login: '/api/v1/auth/login',
  logout: '/api/v1/auth/logout',
  resetPassword: '/api/v1/auth/password/reset',
  sendOtp: '/api/v1/auth/otp/send',
  signup: '/api/v1/auth/signup',
  verifyOtp: '/api/v1/auth/otp/verify',
} as const;

function withSignal(signal?: AbortSignal) {
  return signal ? { signal } : {};
}

export function signup(payload: SignupRequest, signal?: AbortSignal) {
  return postJson<AuthSessionResponse, SignupRequest>(
    authEndpoints.signup,
    payload,
    withSignal(signal),
  );
}

export function login(payload: LoginRequest, signal?: AbortSignal) {
  return postJson<AuthSessionResponse, LoginRequest>(authEndpoints.login, payload, {
    ...withSignal(signal),
  });
}

export function logout() {
  return postJson<SuccessResponse, Record<string, never>>(
    authEndpoints.logout,
    {},
    { auth: true },
  );
}

export function sendOtp(payload: SendOtpRequest, signal?: AbortSignal) {
  return postJson<OtpChallengeResponse, SendOtpRequest>(
    authEndpoints.sendOtp,
    payload,
    withSignal(signal),
  );
}

export function verifyOtp(payload: VerifyOtpRequest, signal?: AbortSignal) {
  return postJson<SuccessResponse, VerifyOtpRequest>(
    authEndpoints.verifyOtp,
    payload,
    withSignal(signal),
  );
}

export function forgotPassword(
  payload: ForgotPasswordRequest,
  signal?: AbortSignal,
) {
  return postJson<OtpChallengeResponse, ForgotPasswordRequest>(
    authEndpoints.forgotPassword,
    payload,
    withSignal(signal),
  );
}

export function resetPassword(
  payload: ResetPasswordRequest,
  signal?: AbortSignal,
) {
  return postJson<SuccessResponse, ResetPasswordRequest>(
    authEndpoints.resetPassword,
    payload,
    withSignal(signal),
  );
}
