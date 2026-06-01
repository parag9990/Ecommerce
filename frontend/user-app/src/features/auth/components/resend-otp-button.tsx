import { useEffect, useState } from 'react';

import { Button } from '../../../components/ui/button';

type ResendOtpButtonProps = {
  disabled?: boolean;
  onResend: () => Promise<void>;
};

const RESEND_COOLDOWN_SECONDS = 30;

export function ResendOtpButton({
  disabled = false,
  onResend,
}: ResendOtpButtonProps) {
  const [cooldown, setCooldown] = useState(0);
  const [isSubmitting, setIsSubmitting] = useState(false);

  useEffect(() => {
    if (cooldown <= 0) {
      return undefined;
    }

    const timerId = window.setTimeout(() => {
      setCooldown((current) => Math.max(current - 1, 0));
    }, 1000);

    return () => {
      window.clearTimeout(timerId);
    };
  }, [cooldown]);

  async function handleClick() {
    setIsSubmitting(true);

    try {
      await onResend();
      setCooldown(RESEND_COOLDOWN_SECONDS);
    } catch {
      // The parent page owns the user-facing error message.
    } finally {
      setIsSubmitting(false);
    }
  }

  const isDisabled = disabled || isSubmitting || cooldown > 0;
  const buttonLabel =
    cooldown > 0 ? `Resend OTP in ${cooldown}s` : 'Resend OTP';

  return (
    <Button
      disabled={isDisabled}
      fullWidth
      onClick={() => {
        void handleClick();
      }}
      type="button"
      variant="secondary"
    >
      {isSubmitting ? 'Sending...' : buttonLabel}
    </Button>
  );
}
