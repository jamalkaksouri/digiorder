#!/bin/bash
for i in {1..100}; do
  curl -s -X POST http://localhost:5582/api/v1/products \
    -H "Content-Type: application/json" \
    -d '{
      "name": "Product '$i'",
      "brand": "Brand '$i'",
      "dosage_form_id": 1,
      "strength": "'$i'mg",
      "unit": "tablet",
      "category_id": 1,
      "description": "Test product '$i'"
    }' > /dev/null
  echo "Created product $i"
done