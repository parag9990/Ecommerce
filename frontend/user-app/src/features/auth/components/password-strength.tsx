import { passwordRequirements } from '../schemas';

type PasswordStrengthProps = {
  password: string;
};

export function PasswordStrength({ password }: PasswordStrengthProps) {
  const passedCount = passwordRequirements.filter((requirement) =>
    requirement.test(password),
  ).length;
  const strengthLabel =
    passedCount === passwordRequirements.length ? 'Strong' : 'Needs work';

  return (
    <div className="rounded-md border border-slate-200 bg-slate-50 px-3 py-2">
      <div className="mb-2 flex items-center justify-between gap-3 text-xs font-medium">
        <span className="text-slate-600">Password strength</span>
        <span
          className={
            passedCount === passwordRequirements.length
              ? 'text-emerald-700'
              : 'text-slate-600'
          }
        >
          {strengthLabel}
        </span>
      </div>

      <ul className="grid gap-1 text-xs text-slate-600">
        {passwordRequirements.map((requirement) => {
          const passed = requirement.test(password);

          return (
            <li className="flex items-center gap-2" key={requirement.id}>
              <span
                aria-hidden="true"
                className={[
                  'h-1.5 w-1.5 rounded-full',
                  passed ? 'bg-emerald-500' : 'bg-slate-300',
                ].join(' ')}
              />
              <span className={passed ? 'text-slate-800' : undefined}>
                {requirement.label}
              </span>
            </li>
          );
        })}
      </ul>
    </div>
  );
}
