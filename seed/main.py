import random
from datetime import datetime, timedelta

import psycopg2

DB_CONFIG = {
    "dbname": "imp_db",
    "user": "postgres",
    "password": "123456",
    "host": "localhost",
    "port": "5432",
}

ITEM_COUNT = 100
DAYS = 7
MOVEMENTS_PER_DAY = 2000

conn = psycopg2.connect(**DB_CONFIG)
cur = conn.cursor()

# =========================
# 1. INSERT ITEMS (100 items)
# =========================
print("Seeding items...")

item_ids = []

for i in range(ITEM_COUNT):
    cur.execute(
        """
        INSERT INTO inventory_items (name, sku, current_stock, low_stock_threshold, created_at, updated_at)
        VALUES (%s, %s, %s, %s, NOW(), NOW())
        RETURNING id
    """,
        (f"Item-{i + 1}", f"SKU-{i + 1:04d}", 0, 10),
    )

    item_ids.append(cur.fetchone()[0])

conn.commit()

# =========================
# 2. INSERT MOVEMENTS
# =========================
# print("Seeding movements...")

# movement_types = ["IN", "OUT"]

# for day in range(DAYS):
#     base_date = datetime.now() - timedelta(days=day)

#     rows = []

#     for i in range(MOVEMENTS_PER_DAY):
#         item_id = random.choice(item_ids)
#         mtype = random.choice(movement_types)

#         qty = random.randint(1, 20)

#         rows.append(
#             (
#                 f"EXT-{day}-{i}-{item_id}",
#                 item_id,
#                 mtype,
#                 qty,
#                 base_date,
#                 base_date,
#                 "seed",
#             )
#         )

#     cur.executemany(
#         """
#         INSERT INTO inventory_movements (
#             external_id,
#             item_id,
#             movement_type,
#             quantity,
#             created_at,
#             updated_at,
#             note
#         )
#         VALUES (%s, %s, %s, %s, %s, %s, %s)
#     """,
#         rows,
#     )

#     conn.commit()
#     print(f"Inserted day {day + 1}/{DAYS}")

cur.close()
conn.close()

print("DONE: 100 items + 14k movements seeded")
