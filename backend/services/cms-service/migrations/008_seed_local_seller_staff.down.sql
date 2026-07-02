DELETE FROM seller_staff
WHERE staff_id = 'staff_local_seller_owner'
   OR (seller_id = 'seller_local_demo' AND user_id = 'user_local_seller');
