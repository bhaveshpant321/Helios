erDiagram
    users {
        bigint id PK
        varchar username UK
        varchar email UK
        varchar password_hash
        timestamptz created_at
    }

    assets {
        int id PK
        varchar ticker_symbol UK
        varchar name
        int decimals
    }

    trading_pairs {
        int id PK
        int base_asset_id FK
        int quote_asset_id FK
        varchar symbol UK
    }

    accounts {
        bigint user_id PK, FK
        int asset_id PK, FK
        decimal balance
    }

    orders {
        bigint id PK
        bigint user_id FK
        int trading_pair_id FK
        order_side side
        order_type type
        order_status status
        decimal price
        decimal quantity
        decimal filled_quantity
        timestamptz created_at
    }

    trades {
        bigint id PK
        bigint maker_order_id FK
        bigint taker_order_id FK
        int trading_pair_id FK
        decimal price
        decimal quantity
        timestamptz executed_at
    }

    fees {
        bigint id PK
        bigint trade_id FK
        bigint user_id FK
        varchar fee_type
        decimal amount
    }

    users ||--o{ accounts : "has"
    assets ||--o{ accounts : "of"
    users ||--o{ orders : "places"
    trading_pairs ||--o{ orders : "for"
    trading_pairs ||--o{ trades : "for"
    assets ||--|{ trading_pairs : "is base"
    assets ||--|{ trading_pairs : "is quote"
    orders }|--o{ trades : "is maker"
    orders }|--o{ trades : "is taker"
    trades ||--o{ fees : "generates"
    users ||--o{ fees : "pays"