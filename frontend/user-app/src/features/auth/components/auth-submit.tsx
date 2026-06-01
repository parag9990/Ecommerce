import { Button } from '../../../components/ui/button';

type AuthSubmitProps = {
  children: string;
  isSubmitting: boolean;
};

export function AuthSubmit({ children, isSubmitting }: AuthSubmitProps) {
  return (
    <Button disabled={isSubmitting} fullWidth type="submit">
      {isSubmitting ? 'Please wait...' : children}
    </Button>
  );
}
