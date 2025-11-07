TOKEN=$1

if [ -z "$TOKEN" ]; then
    echo "Usage: ./seed_data.sh <JWT_TOKEN>"
    exit 1
fi

BASE_URL="http://localhost:5582/api/v1"

# Create categories
echo "Creating categories..."
curl -X POST $BASE_URL/categories \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "دارویی"}'

curl -X POST $BASE_URL/categories \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "آرایشی"}'

# Create dosage forms
echo "Creating dosage forms..."
curl -X POST $BASE_URL/dosage_forms \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "قرص"}'

curl -X POST $BASE_URL/dosage_forms \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "شربت"}'

# Create products
echo "Creating products..."
curl -X POST $BASE_URL/products \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "آموکسی سیلین",
    "brand": "Bayer",
    "dosage_form_id": 1,
    "strength": "500mg",
    "unit": "بسته",
    "category_id": 1,
    "description": "آنتی بیوتیک گسترده الطیف"
  }'

echo "Sample data created!"