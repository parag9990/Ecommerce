-- Roll back OTP challenge account ownership constraint.

USE auth_db;

ALTER TABLE otp_challenges
  DROP FOREIGN KEY fk_otp_challenges_account;
