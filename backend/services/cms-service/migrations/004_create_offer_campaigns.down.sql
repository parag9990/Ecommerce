ALTER TABLE coupon_redemptions
  DROP FOREIGN KEY fk_coupon_redemptions_campaign,
  DROP KEY idx_coupon_redemptions_campaign_user,
  DROP KEY idx_coupon_redemptions_campaign_created,
  DROP COLUMN campaign_id;

DROP TABLE IF EXISTS campaigns;
