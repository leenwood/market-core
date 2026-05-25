package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		log.Fatal("DATABASE_DSN is required")
	}

	ctx := context.Background()
	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		log.Fatalf("ping: %v", err)
	}

	log.Println("seeding categories...")
	cats, err := seedCategories(ctx, db)
	if err != nil {
		log.Fatalf("seed categories: %v", err)
	}

	log.Println("seeding products...")
	if err := seedProducts(ctx, db, cats); err != nil {
		log.Fatalf("seed products: %v", err)
	}

	log.Println("seeding search queries...")
	if err := seedSearchQueries(ctx, db); err != nil {
		log.Fatalf("seed search queries: %v", err)
	}

	log.Println("done.")
}

// ── Categories ────────────────────────────────────────────────────────────────

type category struct {
	id       uuid.UUID
	name     string
	slug     string
	parentID *uuid.UUID
}

func seedCategories(ctx context.Context, db *pgxpool.Pool) (map[string]uuid.UUID, error) {
	defs := []category{
		{name: "Электроника", slug: "electronics"},
		{name: "Смартфоны", slug: "smartphones"},
		{name: "Ноутбуки", slug: "laptops"},
		{name: "Планшеты", slug: "tablets"},
		{name: "Аксессуары", slug: "accessories"},
		{name: "Наушники", slug: "headphones"},
	}

	ids := make(map[string]uuid.UUID, len(defs))
	for i := range defs {
		defs[i].id = uuid.New()
		ids[defs[i].slug] = defs[i].id
	}

	// set parent_id for sub-categories
	electronics := ids["electronics"]
	for i := range defs {
		if defs[i].slug != "electronics" {
			defs[i].parentID = &electronics
		}
	}

	for _, c := range defs {
		_, err := db.Exec(ctx, `
			INSERT INTO categories (id, name, slug, parent_id, sort_order, created_at, updated_at)
			VALUES ($1, $2, $3, $4, 0, NOW(), NOW())
			ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name`,
			c.id, c.name, c.slug, c.parentID,
		)
		if err != nil {
			return nil, fmt.Errorf("insert category %q: %w", c.slug, err)
		}
	}

	log.Printf("  inserted %d categories", len(defs))
	return ids, nil
}

// ── Products ──────────────────────────────────────────────────────────────────

type product struct {
	name        string
	description string
	categoryID  uuid.UUID
	brand       string
	price       float64
	inStock     bool
	rating      float64
	ratingCount int
	viewCount   int64
	attributes  map[string]any
}

func seedProducts(ctx context.Context, db *pgxpool.Pool, cats map[string]uuid.UUID) error {
	products := []product{
		// Смартфоны
		{
			name:        "iPhone 15 Pro Max",
			description: "Флагманский смартфон Apple с чипом A17 Pro и титановым корпусом",
			categoryID:  cats["smartphones"],
			brand:       "Apple",
			price:       149990,
			inStock:     true,
			rating:      4.9,
			ratingCount: 1842,
			viewCount:   54200,
			attributes:  map[string]any{"color": "Natural Titanium", "storage": "256GB", "ram": "8GB", "screen": "6.7 inch"},
		},
		{
			name:        "iPhone 15",
			description: "Смартфон Apple с Dynamic Island и камерой 48 МП",
			categoryID:  cats["smartphones"],
			brand:       "Apple",
			price:       89990,
			inStock:     true,
			rating:      4.8,
			ratingCount: 3201,
			viewCount:   41300,
			attributes:  map[string]any{"color": "Blue", "storage": "128GB", "ram": "6GB", "screen": "6.1 inch"},
		},
		{
			name:        "Samsung Galaxy S24 Ultra",
			description: "Флагман Samsung с встроенным стилусом S Pen и AI-функциями",
			categoryID:  cats["smartphones"],
			brand:       "Samsung",
			price:       129990,
			inStock:     true,
			rating:      4.8,
			ratingCount: 2103,
			viewCount:   38900,
			attributes:  map[string]any{"color": "Titanium Black", "storage": "256GB", "ram": "12GB", "screen": "6.8 inch"},
		},
		{
			name:        "Samsung Galaxy S24",
			description: "Компактный флагман Samsung с чипом Snapdragon 8 Gen 3",
			categoryID:  cats["smartphones"],
			brand:       "Samsung",
			price:       79990,
			inStock:     true,
			rating:      4.7,
			ratingCount: 1540,
			viewCount:   28700,
			attributes:  map[string]any{"color": "Onyx Black", "storage": "128GB", "ram": "8GB", "screen": "6.2 inch"},
		},
		{
			name:        "Google Pixel 8 Pro",
			description: "Смартфон Google с чипом Tensor G3 и продвинутой камерой",
			categoryID:  cats["smartphones"],
			brand:       "Google",
			price:       99990,
			inStock:     true,
			rating:      4.7,
			ratingCount: 892,
			viewCount:   19200,
			attributes:  map[string]any{"color": "Bay", "storage": "128GB", "ram": "12GB", "screen": "6.7 inch"},
		},
		{
			name:        "Xiaomi 14 Pro",
			description: "Флагман Xiaomi с камерой Leica и зарядкой 120 Вт",
			categoryID:  cats["smartphones"],
			brand:       "Xiaomi",
			price:       74990,
			inStock:     false,
			rating:      4.6,
			ratingCount: 743,
			viewCount:   15800,
			attributes:  map[string]any{"color": "White", "storage": "256GB", "ram": "12GB", "screen": "6.73 inch"},
		},
		{
			name:        "OnePlus 12",
			description: "Флагман OnePlus с Hasselblad камерой и зарядкой 100 Вт",
			categoryID:  cats["smartphones"],
			brand:       "OnePlus",
			price:       64990,
			inStock:     true,
			rating:      4.6,
			ratingCount: 521,
			viewCount:   12300,
			attributes:  map[string]any{"color": "Silky Black", "storage": "256GB", "ram": "16GB", "screen": "6.82 inch"},
		},

		// Ноутбуки
		{
			name:        "MacBook Pro 14 M3 Pro",
			description: "Профессиональный ноутбук Apple с чипом M3 Pro и дисплеем Liquid Retina XDR",
			categoryID:  cats["laptops"],
			brand:       "Apple",
			price:       219990,
			inStock:     true,
			rating:      4.9,
			ratingCount: 934,
			viewCount:   31200,
			attributes:  map[string]any{"cpu": "M3 Pro", "ram": "18GB", "storage": "512GB SSD", "screen": "14.2 inch", "os": "macOS"},
		},
		{
			name:        "MacBook Air 15 M2",
			description: "Тонкий и лёгкий ноутбук Apple с чипом M2 и большим дисплеем",
			categoryID:  cats["laptops"],
			brand:       "Apple",
			price:       149990,
			inStock:     true,
			rating:      4.8,
			ratingCount: 1203,
			viewCount:   27400,
			attributes:  map[string]any{"cpu": "M2", "ram": "8GB", "storage": "256GB SSD", "screen": "15.3 inch", "os": "macOS"},
		},
		{
			name:        "Dell XPS 15",
			description: "Ноутбук Dell с OLED-дисплеем и процессором Intel Core Ultra 9",
			categoryID:  cats["laptops"],
			brand:       "Dell",
			price:       189990,
			inStock:     true,
			rating:      4.7,
			ratingCount: 678,
			viewCount:   18900,
			attributes:  map[string]any{"cpu": "Intel Core Ultra 9", "ram": "32GB", "storage": "1TB SSD", "screen": "15.6 inch", "os": "Windows 11"},
		},
		{
			name:        "Lenovo ThinkPad X1 Carbon",
			description: "Бизнес-ноутбук с лёгким карбоновым корпусом и долгой батареей",
			categoryID:  cats["laptops"],
			brand:       "Lenovo",
			price:       159990,
			inStock:     true,
			rating:      4.7,
			ratingCount: 445,
			viewCount:   14200,
			attributes:  map[string]any{"cpu": "Intel Core i7-1365U", "ram": "16GB", "storage": "512GB SSD", "screen": "14 inch", "os": "Windows 11 Pro"},
		},
		{
			name:        "ASUS ROG Zephyrus G16",
			description: "Игровой ноутбук с RTX 4080 и дисплеем 240 Гц",
			categoryID:  cats["laptops"],
			brand:       "ASUS",
			price:       229990,
			inStock:     false,
			rating:      4.8,
			ratingCount: 312,
			viewCount:   22600,
			attributes:  map[string]any{"cpu": "Intel Core i9-14900H", "ram": "32GB", "storage": "2TB SSD", "gpu": "RTX 4080", "screen": "16 inch", "os": "Windows 11"},
		},
		{
			name:        "HP Spectre x360 14",
			description: "Трансформер HP с OLED-дисплеем и поддержкой стилуса",
			categoryID:  cats["laptops"],
			brand:       "HP",
			price:       139990,
			inStock:     true,
			rating:      4.6,
			ratingCount: 289,
			viewCount:   11800,
			attributes:  map[string]any{"cpu": "Intel Core Ultra 7", "ram": "16GB", "storage": "512GB SSD", "screen": "14 inch", "os": "Windows 11"},
		},

		// Планшеты
		{
			name:        "iPad Pro 13 M4",
			description: "Профессиональный планшет Apple с чипом M4 и дисплеем OLED",
			categoryID:  cats["tablets"],
			brand:       "Apple",
			price:       159990,
			inStock:     true,
			rating:      4.9,
			ratingCount: 567,
			viewCount:   19800,
			attributes:  map[string]any{"storage": "256GB", "connectivity": "WiFi + Cellular", "screen": "13 inch", "os": "iPadOS 17"},
		},
		{
			name:        "iPad Air 11 M2",
			description: "Универсальный планшет Apple с чипом M2",
			categoryID:  cats["tablets"],
			brand:       "Apple",
			price:       79990,
			inStock:     true,
			rating:      4.8,
			ratingCount: 892,
			viewCount:   24100,
			attributes:  map[string]any{"storage": "128GB", "connectivity": "WiFi", "screen": "11 inch", "os": "iPadOS 17"},
		},
		{
			name:        "Samsung Galaxy Tab S9 Ultra",
			description: "Флагманский планшет Samsung с AMOLED-дисплеем 14.6 дюйма",
			categoryID:  cats["tablets"],
			brand:       "Samsung",
			price:       109990,
			inStock:     true,
			rating:      4.7,
			ratingCount: 431,
			viewCount:   16700,
			attributes:  map[string]any{"storage": "256GB", "ram": "12GB", "connectivity": "WiFi + 5G", "screen": "14.6 inch", "os": "Android 13"},
		},
		{
			name:        "Xiaomi Pad 6 Pro",
			description: "Планшет Xiaomi с процессором Snapdragon 8+ Gen 1 и зарядкой 67 Вт",
			categoryID:  cats["tablets"],
			brand:       "Xiaomi",
			price:       44990,
			inStock:     true,
			rating:      4.5,
			ratingCount: 654,
			viewCount:   12300,
			attributes:  map[string]any{"storage": "256GB", "ram": "8GB", "connectivity": "WiFi", "screen": "11 inch", "os": "Android 13"},
		},

		// Наушники
		{
			name:        "AirPods Pro 2",
			description: "Беспроводные наушники Apple с активным шумоподавлением и чипом H2",
			categoryID:  cats["headphones"],
			brand:       "Apple",
			price:       24990,
			inStock:     true,
			rating:      4.9,
			ratingCount: 4201,
			viewCount:   67800,
			attributes:  map[string]any{"type": "TWS", "anc": true, "battery": "6h", "case_battery": "30h", "color": "White"},
		},
		{
			name:        "Sony WH-1000XM5",
			description: "Накладные наушники Sony с лучшим в классе шумоподавлением",
			categoryID:  cats["headphones"],
			brand:       "Sony",
			price:       29990,
			inStock:     true,
			rating:      4.9,
			ratingCount: 3872,
			viewCount:   52400,
			attributes:  map[string]any{"type": "Over-ear", "anc": true, "battery": "30h", "color": "Black"},
		},
		{
			name:        "Samsung Galaxy Buds3 Pro",
			description: "Наушники Samsung с адаптивным шумоподавлением и Hi-Fi звуком",
			categoryID:  cats["headphones"],
			brand:       "Samsung",
			price:       17990,
			inStock:     true,
			rating:      4.7,
			ratingCount: 1234,
			viewCount:   23100,
			attributes:  map[string]any{"type": "TWS", "anc": true, "battery": "6h", "case_battery": "30h", "color": "Silver"},
		},
		{
			name:        "Bose QuietComfort Ultra",
			description: "Премиум наушники с иммерсивным пространственным звуком",
			categoryID:  cats["headphones"],
			brand:       "Bose",
			price:       39990,
			inStock:     true,
			rating:      4.8,
			ratingCount: 987,
			viewCount:   18900,
			attributes:  map[string]any{"type": "Over-ear", "anc": true, "battery": "24h", "color": "Black"},
		},
		{
			name:        "Jabra Evolve2 85",
			description: "Профессиональные наушники для бизнеса с 8-микрофонной системой",
			categoryID:  cats["headphones"],
			brand:       "Jabra",
			price:       34990,
			inStock:     false,
			rating:      4.6,
			ratingCount: 432,
			viewCount:   9800,
			attributes:  map[string]any{"type": "Over-ear", "anc": true, "battery": "37h", "color": "Black"},
		},

		// Аксессуары
		{
			name:        "Apple Watch Series 9",
			description: "Умные часы Apple с чипом S9 и функцией Double Tap",
			categoryID:  cats["accessories"],
			brand:       "Apple",
			price:       39990,
			inStock:     true,
			rating:      4.8,
			ratingCount: 2341,
			viewCount:   43200,
			attributes:  map[string]any{"size": "45mm", "color": "Midnight", "material": "Aluminium", "connectivity": "GPS + Cellular"},
		},
		{
			name:        "MagSafe Charger 15W",
			description: "Беспроводное зарядное устройство Apple MagSafe с мощностью 15 Вт",
			categoryID:  cats["accessories"],
			brand:       "Apple",
			price:       4990,
			inStock:     true,
			rating:      4.5,
			ratingCount: 3102,
			viewCount:   31400,
			attributes:  map[string]any{"power": "15W", "connector": "USB-C", "color": "White"},
		},
		{
			name:        "Samsung 45W Super Fast Charger",
			description: "Быстрое зарядное устройство Samsung с кабелем USB-C",
			categoryID:  cats["accessories"],
			brand:       "Samsung",
			price:       2990,
			inStock:     true,
			rating:      4.6,
			ratingCount: 1876,
			viewCount:   19800,
			attributes:  map[string]any{"power": "45W", "connector": "USB-C"},
		},
		{
			name:        "Anker PowerCore 26800",
			description: "Портативный аккумулятор Anker ёмкостью 26800 мАч с двумя USB-A и USB-C",
			categoryID:  cats["accessories"],
			brand:       "Anker",
			price:       5990,
			inStock:     true,
			rating:      4.7,
			ratingCount: 2654,
			viewCount:   28700,
			attributes:  map[string]any{"capacity": "26800mAh", "ports": "2x USB-A + 1x USB-C", "color": "Black"},
		},
		{
			name:        "Baseus 65W GaN Charger",
			description: "Компактный GaN-блок питания Baseus для зарядки ноутбука и телефона",
			categoryID:  cats["accessories"],
			brand:       "Baseus",
			price:       3490,
			inStock:     true,
			rating:      4.5,
			ratingCount: 1102,
			viewCount:   14300,
			attributes:  map[string]any{"power": "65W", "ports": "2x USB-C + 1x USB-A", "color": "White"},
		},
		{
			name:        "Logitech MX Master 3S",
			description: "Беспроводная мышь Logitech с тихим кликом и точным сенсором 8000 DPI",
			categoryID:  cats["accessories"],
			brand:       "Logitech",
			price:       8990,
			inStock:     true,
			rating:      4.9,
			ratingCount: 4321,
			viewCount:   38900,
			attributes:  map[string]any{"connectivity": "Bluetooth + 2.4GHz", "dpi": "8000", "battery": "70 days", "color": "Graphite"},
		},
	}

	for _, p := range products {
		attrs, err := json.Marshal(p.attributes)
		if err != nil {
			return fmt.Errorf("marshal attributes: %w", err)
		}
		_, err = db.Exec(ctx, `
			INSERT INTO products
				(id, name, description, category_id, brand, price, rating, rating_count,
				 in_stock, attributes, view_count, created_at, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$12)
			ON CONFLICT DO NOTHING`,
			uuid.New(), p.name, p.description, p.categoryID, p.brand,
			p.price, p.rating, p.ratingCount, p.inStock, attrs,
			p.viewCount, time.Now(),
		)
		if err != nil {
			return fmt.Errorf("insert product %q: %w", p.name, err)
		}
	}

	log.Printf("  inserted %d products", len(products))
	return nil
}

// ── Search queries ────────────────────────────────────────────────────────────

func seedSearchQueries(ctx context.Context, db *pgxpool.Pool) error {
	queries := []struct {
		query   string
		results int
		count   int
	}{
		{"iphone 15", 8, 142},
		{"macbook pro", 6, 98},
		{"samsung galaxy", 12, 87},
		{"наушники sony", 4, 76},
		{"airpods pro", 3, 231},
		{"ноутбук apple", 9, 65},
		{"xiaomi смартфон", 7, 54},
		{"планшет samsung", 5, 43},
		{"зарядка usb-c", 8, 112},
		{"игровой ноутбук", 4, 38},
	}

	for _, q := range queries {
		for i := 0; i < q.count; i++ {
			createdAt := time.Now().Add(-time.Duration(i*2+1) * time.Hour)
			_, err := db.Exec(ctx, `
				INSERT INTO search_queries (id, query, results_count, created_at)
				VALUES ($1, $2, $3, $4)`,
				uuid.New(), q.query, q.results, createdAt,
			)
			if err != nil {
				return fmt.Errorf("insert search query %q: %w", q.query, err)
			}
		}
	}

	total := 0
	for _, q := range queries {
		total += q.count
	}
	log.Printf("  inserted %d search query records", total)
	return nil
}
