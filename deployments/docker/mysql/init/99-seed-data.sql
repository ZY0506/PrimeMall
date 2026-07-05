-- ============================================================
-- PrimeMall 清库 + 种子数据（简化版，使用 CTE）
-- 使用方法: mysql -h127.0.0.1 -P13306 -uroot -p123456 < 此文件
-- ============================================================

-- ============================================================
-- 第一步：清空所有表
-- ============================================================
DELETE FROM shop_order.after_sale_item;
DELETE FROM shop_order.after_sale;
DELETE FROM shop_order.order_item;
DELETE FROM shop_order.order_info;
DELETE FROM shop_order.cart;

DELETE FROM shop_product.stock_log;
DELETE FROM shop_product.product_sku;
DELETE FROM shop_product.product_spu;
DELETE FROM shop_product.freight_template;
DELETE FROM shop_product.category;

DELETE FROM shop_payment.payment_callback_log;
DELETE FROM shop_payment.refund;
DELETE FROM shop_payment.payment;

DELETE FROM shop_marketing.seckill_preorder;
DELETE FROM shop_marketing.seckill_activity;
DELETE FROM shop_marketing.user_coupon;
DELETE FROM shop_marketing.coupon;

DELETE FROM shop_user.user_login_log;
DELETE FROM shop_user.user_punish_log;
DELETE FROM shop_user.user_address;
DELETE FROM shop_user.user;

DELETE FROM shop_search.es_sync_log;
DELETE FROM shop_search.search_history;
DELETE FROM shop_search.hot_keyword;

DELETE FROM shop_admin.help_article;
DELETE FROM shop_admin.help_category;
DELETE FROM shop_admin.notice;
DELETE FROM shop_admin.banner;
DELETE FROM shop_admin.admin_log;
DELETE FROM shop_admin.role_permission;
DELETE FROM shop_admin.permission;
DELETE FROM shop_admin.role;
DELETE FROM shop_admin.admin;

-- ============================================================
-- 第二步：用户数据 (shop_user)
-- ============================================================

-- 500 个压测用户，密码: Test123456
INSERT INTO shop_user.`user` (id, phone, password, nickname, gender, status, created_at)
WITH RECURSIVE nums(n) AS (
  SELECT 1 UNION ALL SELECT n + 1 FROM nums WHERE n < 500
)
SELECT
  2064000000000000000 + n,
  CONCAT('13800000', LPAD(n, 3, '0')),
  '$2a$10$dYRyqk9VEbdDNaVPcfskYeN74sdQkFb3I2WCLXkPYePmWihk0QyXG',
  CONCAT('测试用户', LPAD(n, 4, '0')),
  FLOOR(RAND() * 3) AS gender,
  1 AS status,
  NOW()
FROM nums;

-- 地址：奇数用户 2 个，偶数用户 1 个，共约 750 个
INSERT INTO shop_user.user_address (id, user_id, tag, receiver_name, receiver_phone, is_default, info)
WITH RECURSIVE nums(n) AS (
  SELECT 1 UNION ALL SELECT n + 1 FROM nums WHERE n < 500
)
SELECT
  (n - 1) + CEILING((n - 1) / 2) + 1 + a.offset,
  2064000000000000000 + n,
  CASE WHEN a.offset = 0 THEN 'HOME' ELSE 'OFFICE' END,
  CONCAT('收货人', LPAD(n, 4, '0')),
  CONCAT('1380000', LPAD(n + a.offset * 100, 4, '0')),
  CASE WHEN a.offset = 0 THEN 1 ELSE 0 END,
  CASE WHEN a.offset = 0
    THEN JSON_OBJECT('province','广东省','city','深圳市','district','南山区','detail','科技园A栋','postal_code','518000')
    ELSE JSON_OBJECT('province','广东省','city','广州市','district','天河区','detail','写字楼B座','postal_code','510000')
  END
FROM nums
CROSS JOIN (SELECT 0 AS offset UNION SELECT 1) a
WHERE NOT (n % 2 = 0 AND a.offset = 1);

-- ============================================================
-- 第三步：商品数据 (shop_product)
-- ============================================================

-- 分类
INSERT INTO shop_product.category (id, parent_id, name, icon, sort, level, status) VALUES
(100, 0, '手机数码',  '/icons/phone.png',    1, 1, 1),
(110, 100, '手机',   '/icons/smartphone.png',1, 2, 1),
(111, 100, '平板电脑','/icons/tablet.png',   2, 2, 1),
(112, 100, '手机配件','/icons/accessories.png',3,2,1),
(120, 110, '苹果手机','',1,3,1), (121,110,'华为手机','',2,3,1),
(122, 110, '小米手机','',3,3,1), (123,110,'其他品牌','',4,3,1),
(200, 0, '电脑办公', '/icons/computer.png',  2, 1, 1),
(210, 200, '笔记本', '',1,2,1), (211,200,'台式机/显示器','',2,2,1),
(212, 200, '电脑外设','',3,2,1),
(300, 0, '影音娱乐', '/icons/audio.png',    3, 1, 1),
(310, 300, '耳机',   '',1,2,1), (311,300,'音箱','',2,2,1),
(400, 0, '智能穿戴', '/icons/wearable.png', 4, 1, 1),
(410, 400, '智能手表','',1,2,1), (411,400,'手环','',2,2,1);

-- SPU
INSERT INTO shop_product.product_spu (id,category_id,name,brand,`desc`,main_pic,status,sales_count,virtual_sales,freight_template_id) VALUES
(1001,120,'iPhone 16 Pro Max','Apple','苹果最新旗舰手机','/img/iphone16.jpg',1,5000,10000,1),
(1002,120,'iPhone 16','Apple','苹果新一代手机','/img/iphone16.jpg',1,3000,8000,1),
(1003,121,'华为 Mate 70 Pro','华为','华为旗舰手机','/img/mate70.jpg',1,4000,9000,1),
(1004,121,'华为 Pura 70','华为','华为影像旗舰','/img/pura70.jpg',1,3500,8500,1),
(1005,122,'小米 15 Pro','小米','小米年度旗舰','/img/mi15pro.jpg',1,4500,9500,1),
(1006,122,'Redmi K80 Pro','Redmi','Redmi性能旗舰','/img/k80pro.jpg',1,6000,11000,1),
(1007,123,'一加 13','一加','一加旗舰手机','/img/oneplus13.jpg',1,2500,7000,1),
(1008,123,'OPPO Find X8','OPPO','OPPO影像旗舰','/img/findx8.jpg',1,2000,6000,1),
(1009,210,'MacBook Pro 16','Apple','苹果高性能笔记本','/img/mbp16.jpg',1,2000,5000,1),
(1010,210,'ThinkPad X1 Carbon','联想','商务轻薄本','/img/x1carbon.jpg',1,1800,4500,1),
(1011,211,'ROG PG27AQDM','华硕','27寸2K OLED电竞显示器','/img/rog27.jpg',1,800,2000,1),
(1012,211,'Dell U2723QE','戴尔','27寸4K IPS专业显示器','/img/dell27.jpg',1,600,1500,1),
(1013,212,'罗技 GPW 三代','罗技','无线游戏鼠标','/img/gpw3.jpg',1,3000,7000,1),
(1014,212,'Keychron Q6','Keychron','机械键盘','/img/keychron.jpg',1,1500,4000,1),
(1015,310,'AirPods Pro 3','Apple','苹果主动降噪耳机','/img/airpods3.jpg',1,5000,12000,1),
(1016,310,'华为 FreeBuds Pro 4','华为','华为旗舰降噪耳机','/img/freebuds4.jpg',1,3500,8000,1),
(1017,311,'小米 Buds 5 Pro','小米','小米旗舰耳机','/img/buds5pro.jpg',1,2800,6000,1),
(1018,411,'华为 GT5 Pro','华为','华为高端智能手表','/img/gt5pro.jpg',1,2200,5000,1),
(1019,410,'Apple Watch Ultra 3','Apple','苹果户外旗舰手表','/img/awu3.jpg',1,1800,4000,1),
(1020,311,'Bose QC Ultra','Bose','Bose旗舰降噪耳机','/img/boseqc.jpg',1,1200,3000,1);

-- SKU（与 tests/k6/config.js 的 SKU_IDS 对应）
INSERT INTO shop_product.product_sku (id,spu_id,spu_name,sku_code,price,market_price,cost_price,stock,locked_stock,spec_data,weight,status) VALUES
(10001,1001,'iPhone 16 Pro Max','IP16PM-B256',999900,1099900,750000,10000,0,'[{"name":"颜色","value":"黑色"},{"name":"存储","value":"256GB"}]',250,1),
(10002,1001,'iPhone 16 Pro Max','IP16PM-W256',999900,1099900,750000,8000,0,'[{"name":"颜色","value":"白色"},{"name":"存储","value":"256GB"}]',250,1),
(10003,1001,'iPhone 16 Pro Max','IP16PM-B512',1099900,1199900,820000,5000,0,'[{"name":"颜色","value":"黑色"},{"name":"存储","value":"512GB"}]',250,1),
(10004,1002,'iPhone 16','IP16-B128',599900,659900,450000,12000,0,'[{"name":"颜色","value":"黑色"},{"name":"存储","value":"128GB"}]',200,1),
(10005,1002,'iPhone 16','IP16-P128',599900,659900,450000,10000,0,'[{"name":"颜色","value":"粉色"},{"name":"存储","value":"128GB"}]',200,1),
(10006,1003,'华为 Mate 70 Pro','M70P-B512',699900,769900,520000,8000,0,'[{"name":"颜色","value":"黑色"},{"name":"存储","value":"512GB"}]',220,1),
(10007,1004,'华为 Pura 70','P70-B256',499900,549900,370000,9000,0,'[{"name":"颜色","value":"黑色"},{"name":"存储","value":"256GB"}]',200,1),
(10008,1005,'小米 15 Pro','M15P-B512',529900,579900,390000,10000,0,'[{"name":"颜色","value":"黑色"},{"name":"存储","value":"512GB"}]',210,1),
(10009,1007,'一加 13','OP13-B512',499900,549900,360000,5,0,'[{"name":"颜色","value":"黑色"},{"name":"存储","value":"512GB"}]',210,1),
(10010,1008,'OPPO Find X8','FX8-B256',429900,479900,310000,6000,0,'[{"name":"颜色","value":"黑色"},{"name":"存储","value":"256GB"}]',200,1),
(10011,1009,'MacBook Pro 16 M4 Max','MBP16-M4-64',2999900,3299900,2200000,2000,0,'[{"name":"颜色","value":"深空黑"},{"name":"配置","value":"M4 Max 64GB 1TB"}]',2500,1),
(10012,1009,'MacBook Pro 16 M4 Pro','MBP16-M4P-24',1999900,2199900,1500000,3000,0,'[{"name":"颜色","value":"银色"},{"name":"配置","value":"M4 Pro 24GB 512GB"}]',2400,1),
(10013,1010,'ThinkPad X1 Carbon','X1C13-I7H',1599900,1799900,1200000,2000,0,'[{"name":"颜色","value":"黑色"},{"name":"配置","value":"i7-1365U 16GB 512GB"}]',1800,1),
(10014,1012,'Dell U2723QE','U2723QE-4K',459900,509900,320000,1,0,'[{"name":"颜色","value":"黑色"},{"name":"尺寸","value":"27寸 4K IPS"}]',4500,1),
(10015,1011,'ROG PG27AQDM','PG27Q-OLED',599900,659900,420000,3000,0,'[{"name":"颜色","value":"黑色"},{"name":"尺寸","value":"27寸 2K OLED 240Hz"}]',5000,1),
(10016,1013,'罗技 GPW 三代 黑色','GPW3-BK',29900,39900,18000,20000,0,'[{"name":"颜色","value":"黑色"}]',350,1),
(10017,1013,'罗技 GPW 三代 白色','GPW3-WH',29900,39900,18000,20000,0,'[{"name":"颜色","value":"白色"}]',350,1),
(10018,1014,'Keychron Q6','KQ6-BK',109900,129900,70000,5000,0,'[{"name":"颜色","value":"黑色"},{"name":"轴体","value":"红轴"}]',1500,1),
(10019,1015,'AirPods Pro 3 MagSafe','APP3-MG',189900,209900,130000,2,0,'[{"name":"颜色","value":"白色"},{"name":"充电","value":"MagSafe"}]',80,1),
(10020,1015,'AirPods Pro 3 USB-C','APP3-USBC',179900,199900,125000,5000,0,'[{"name":"颜色","value":"白色"},{"name":"充电","value":"USB-C"}]',80,1),
(10021,1016,'华为 FreeBuds Pro 4','FB4-BK',149900,169900,100000,6000,0,'[{"name":"颜色","value":"黑色"}]',60,1),
(10022,1017,'小米 Buds 5 Pro 黑金','MB5P-BK',99900,119900,70000,2,0,'[{"name":"颜色","value":"黑金"}]',55,1),
(10023,1017,'小米 Buds 5 Pro 月光白','MB5P-WH',99900,119900,70000,2,0,'[{"name":"颜色","value":"月光白"}]',55,1),
(10024,1018,'华为 GT5 Pro 黑色','GT5P-BK',299900,339900,210000,5000,0,'[{"name":"颜色","value":"黑色"},{"name":"表带","value":"氟橡胶"}]',120,1),
(10025,1018,'华为 GT5 Pro 钛金色','GT5P-TI',329900,369900,230000,3000,0,'[{"name":"颜色","value":"钛金色"},{"name":"表带","value":"真皮"}]',120,1),
(10026,1019,'Apple Watch Ultra 3','AWU3-NAT',699900,769900,500000,2000,0,'[{"name":"颜色","value":"自然钛"},{"name":"表带","value":"野径回环"}]',150,1),
(10027,1020,'Bose QC Ultra 黑色','BQCU-BK',399900,449900,280000,3,0,'[{"name":"颜色","value":"黑色"}]',300,1),
(10028,1020,'Bose QC Ultra 白色','BQCU-WH',399900,449900,280000,3600,0,'[{"name":"颜色","value":"白色"}]',300,1),
(10029,1006,'Redmi K80 Pro 512G 黑色','RK80P-B512',399900,439900,300000,10000,0,'[{"name":"颜色","value":"黑色"},{"name":"存储","value":"512GB"}]',210,1),
(10030,1006,'Redmi K80 Pro 512G 白色','RK80P-W512',399900,439900,300000,8000,0,'[{"name":"颜色","value":"白色"},{"name":"存储","value":"512GB"}]',210,1),
(10031,1006,'Redmi K80 Pro 1T 黑色','RK80P-B1T',439900,479900,330000,2,0,'[{"name":"颜色","value":"黑色"},{"name":"存储","value":"1TB"}]',210,1),
(10032,1003,'华为 Mate 70 Pro 白色','M70P-W512',699900,769900,520000,6000,0,'[{"name":"颜色","value":"白色"},{"name":"存储","value":"512GB"}]',220,1),
(10033,1004,'华为 Pura 70 白色','P70-W256',499900,549900,370000,7000,0,'[{"name":"颜色","value":"白色"},{"name":"存储","value":"256GB"}]',200,1),
(10034,1005,'小米 15 Pro 白色','M15P-W512',529900,579900,390000,8000,0,'[{"name":"颜色","value":"白色"},{"name":"存储","value":"512GB"}]',210,1),
(10035,1013,'罗技 GPW 三代 黑色 低库存','GPW3-BK-L',29900,39900,18000,3,0,'[{"name":"颜色","value":"黑色"},{"name":"版本","value":"竞测"}]',350,1),
(10036,1013,'罗技 GPW 三代 白色 低库存','GPW3-WH-L',29900,39900,18000,3,0,'[{"name":"颜色","value":"白色"},{"name":"版本","value":"竞测"}]',350,1);

-- 运费模板
INSERT INTO shop_product.freight_template (id,name,type,default_fee,default_quantity,extra_fee,free_threshold_amount,free_threshold_quantity,is_default,status) VALUES
(1,'默认运费模板',1,800,1,200,9900,0,1,1);

-- ============================================================
-- 第四步：营销数据 (shop_marketing)
-- ============================================================

INSERT INTO shop_marketing.coupon (id,name,type,threshold_amount,reduce_amount,discount_rate,max_discount_amount,total_quantity,used_quantity,per_user_limit,start_time,end_time,status,description) VALUES
(1,'满100减10',  1,10000,1000,0,0,10000,0,3,NOW(),DATE_ADD(NOW(),INTERVAL 30 DAY),1,'全场通用满100减10'),
(2,'满500减60',  1,50000,6000,0,0,5000,0,3,NOW(),DATE_ADD(NOW(),INTERVAL 30 DAY),1,'全场通用满500减60'),
(3,'满1000减150',1,100000,15000,0,0,3000,0,2,NOW(),DATE_ADD(NOW(),INTERVAL 30 DAY),1,'全场通用满1000减150'),
(4,'8折券最高减200',2,0,0,8000,20000,5000,0,2,NOW(),DATE_ADD(NOW(),INTERVAL 30 DAY),1,'全场通用8折优惠'),
(5,'无门槛5元券',3,0,500,0,0,20000,0,1,NOW(),DATE_ADD(NOW(),INTERVAL 30 DAY),1,'全场无门槛立减5元');

-- 每用户发放 1 张随机优惠券
INSERT INTO shop_marketing.user_coupon (coupon_id,user_id,order_sn,status,source,expire_time)
WITH RECURSIVE nums(n) AS (SELECT 1 UNION ALL SELECT n+1 FROM nums WHERE n<500)
SELECT 1+FLOOR(RAND()*5), 2064000000000000000+n, '', 0, 'ACTIVITY', DATE_ADD(NOW(),INTERVAL 30 DAY)
FROM nums;

-- ============================================================
-- 第五步：管理员 (shop_admin)
-- ============================================================

INSERT INTO shop_admin.admin (id,username,password,real_name,status) VALUES
(1,'admin','$2a$10$o/vY0Ep9VFZOx/49JeatjOL03fknAwbmBt2/4HmWuS6u9LZ7DxSjq','系统管理员',1);

INSERT INTO shop_admin.role (id,name,code,remark) VALUES
(1,'超级管理员','ROLE_SUPER_ADMIN','拥有全部权限'),
(2,'运营人员','ROLE_OPERATOR','可管理商品、订单、营销'),
(3,'客服人员','ROLE_CS','可查看订单、处理售后');

-- ============================================================
-- 第六步：搜索 (shop_search)
-- ============================================================

INSERT INTO shop_search.hot_keyword (keyword,search_count,sort,status) VALUES
('iPhone',15000,1,1),('华为手机',12000,2,1),('小米手机',10000,3,1),
('耳机',8000,4,1),('笔记本',7500,5,1),('智能手表',6000,6,1),
('键盘',5000,7,1),('显示器',4500,8,1);

-- ============================================================
-- 第七步：CMS (shop_admin)
-- ============================================================

INSERT INTO shop_admin.banner (title,image_url,type,target_id,sort,status,start_time,end_time) VALUES
('iPhone 16 Pro Max 首发','/banner/iphone16.jpg',1,1001,1,1,NOW(),DATE_ADD(NOW(),INTERVAL 30 DAY)),
('华为 Mate 70 Pro 旗舰新品','/banner/mate70.jpg',1,1003,2,1,NOW(),DATE_ADD(NOW(),INTERVAL 30 DAY)),
('小米 15 Pro 年度旗舰','/banner/mi15pro.jpg',1,1005,3,1,NOW(),DATE_ADD(NOW(),INTERVAL 30 DAY));

INSERT INTO shop_admin.notice (title,content,type,priority,status,publish_time) VALUES
('PrimeMall 系统上线','<p>欢迎来到 PrimeMall！</p>',1,3,1,NOW());
