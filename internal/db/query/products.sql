CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    
    -- نام دارو یا کالا (اجباری)
    name TEXT NOT NULL,
    
    -- برند (اختیاری)
    brand TEXT,
    
    -- شکل دارویی (مثل قرص، شربت، آمپول) → همیشه باید مشخص باشه
    dosage_form_id INT NOT NULL REFERENCES dosage_forms(id),
    
    -- دوز دارو (مثلاً 500mg) → اختیاری
    strength TEXT,
    
    -- واحد (مثلاً بسته، شیشه، ویال) → اختیاری
    unit TEXT,
    
    -- دسته‌بندی (دارویی، آرایشی، بهداشتی) → همیشه باید مشخص باشه
    category_id INT NOT NULL REFERENCES categories(id),
    
    -- توضیحات بیشتر (اختیاری)
    description TEXT,
    
    -- تاریخ ایجاد رکورد
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL
);
