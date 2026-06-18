import { Button } from '../../../components/ui/button';

type PaginationProps = {
  onPageChange: (page: number) => void;
  page: number;
  pageSize: number;
  total: number;
};

export function Pagination({
  onPageChange,
  page,
  pageSize,
  total,
}: PaginationProps) {
  const totalPages = Math.max(1, Math.ceil(total / pageSize));

  if (totalPages <= 1) {
    return null;
  }

  return (
    <nav
      aria-label="Product pagination"
      className="flex items-center justify-between gap-3"
    >
      <Button
        disabled={page <= 1}
        onClick={() => {
          onPageChange(page - 1);
        }}
        variant="secondary"
      >
        Previous
      </Button>

      <span className="text-sm text-slate-600">
        Page {page} of {totalPages}
      </span>

      <Button
        disabled={page >= totalPages}
        onClick={() => {
          onPageChange(page + 1);
        }}
        variant="secondary"
      >
        Next
      </Button>
    </nav>
  );
}
