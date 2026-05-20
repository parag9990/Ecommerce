-- Harden OTP challenge ownership when an account link is present.

USE auth_db;

ALTER TABLE otp_challenges
  ADD CONSTRAINT fk_otp_challenges_account
    FOREIGN KEY (account_id) REFERENCES auth_accounts(account_id);
