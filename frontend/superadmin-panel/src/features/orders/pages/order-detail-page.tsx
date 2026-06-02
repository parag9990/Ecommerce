import { ArrowLeft } from "lucide-react";
import { Link, useParams } from "react-router-dom";

import { DataState } from "../../../components/ui/data-state";
import { PermissionDenied } from "../../../components/ui/permission-denied";
import { DisputePanel } from "../components/dispute-panel";
import { ManualReviewDrawer } from "../components/manual-review-drawer";
import { OrderItemsTable } from "../components/order-items-table";
import { OrderPaymentSummary } from "../components/order-payment-summary";
import { OrderShipmentPanel } from "../components/order-shipment-panel";
import { OrderStatusTimeline } from "../components/order-status-timeline";
import { OrderSummaryPanel } from "../components/order-summary-panel";
import { useAdminOrder } from "../hooks/use-admin-order";
import { useOrderPermissions } from "../permissions";

export function OrderDetailPage() {
  const { orderId = "" } = useParams();
  const permissions = useOrderPermissions();
  const orderQuery = useAdminOrder(orderId);

  if (!permissions.canViewOrderDetail) {
    return <PermissionDenied compact />;
  }

  if (!orderId) {
    return (
      <DataState
        tone="danger"
        title="Order not found"
        description="The route did not include a valid order id."
      />
    );
  }

  if (orderQuery.isLoading) {
    return <DataState title="Loading order" description="Fetching the selected order detail." />;
  }

  if (orderQuery.error) {
    return (
      <DataState
        tone="danger"
        title="Unable to load order"
        description={orderQuery.error instanceof Error ? orderQuery.error.message : "The order request failed."}
        action={
          <button
            type="button"
            onClick={() => void orderQuery.refetch()}
            className="h-9 rounded-md border border-red-200 bg-white px-3 font-medium text-red-800 hover:bg-red-100"
          >
            Retry
          </button>
        }
      />
    );
  }

  if (!orderQuery.data) {
    return <DataState tone="danger" title="Order not found" description="No order matched this id." />;
  }

  const { order, status_history: statusHistory, shipments } = orderQuery.data;
  const payment = orderQuery.data.payment ?? order.payment;

  return (
    <section className="min-h-[calc(100vh-6.5rem)] overflow-hidden rounded-lg border border-slate-200 bg-slate-50">
      <header className="border-b border-slate-200 bg-white px-4 py-4">
        <Link
          to="/admin/orders"
          className="mb-3 inline-flex items-center gap-2 text-sm font-medium text-slate-600 hover:text-slate-950"
        >
          <ArrowLeft size={16} aria-hidden="true" />
          Back to orders
        </Link>
        <div className="min-w-0">
          <h1 className="truncate text-xl font-semibold text-slate-950">Order {order.order_id}</h1>
          <p className="mt-1 break-all text-sm text-slate-600">Buyer {order.user_id}</p>
        </div>
      </header>

      <div className="grid gap-4 p-4 xl:grid-cols-[minmax(0,1fr)_360px]">
        <main className="min-w-0 space-y-4">
          <OrderItemsTable items={order.items} />
          <OrderStatusTimeline events={statusHistory} />
          <DisputePanel orderId={order.order_id} canView={permissions.canViewOrderDisputes} />
        </main>

        <div className="space-y-4">
          <OrderSummaryPanel order={order} />
          <OrderShipmentPanel shipments={shipments} />
          <OrderPaymentSummary payment={payment} />
          <ManualReviewDrawer
            orderId={order.order_id}
            reviewStatus={order.review_status ?? "none"}
            roles={permissions.roles}
          />
        </div>
      </div>
    </section>
  );
}
