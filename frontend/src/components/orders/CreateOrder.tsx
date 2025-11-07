import { useState } from "react";
import { useQuery, useMutation } from "@tanstack/react-query";
import { searchProducts, createOrder } from "@/api";

interface OrderItem {
  product_id: string;
  requested_qty: number;
  unit: string;
  note?: string;
}

export const CreateOrder = () => {
  const [cart, setCart] = useState<OrderItem[]>([]);
  const [search, setSearch] = useState("");

  const { data: products } = useQuery({
    queryKey: ["products", search],
    queryFn: () => searchProducts(search),
  });

  const createOrderMutation = useMutation({
    mutationFn: createOrder,
    onSuccess: () => {
      alert("Order created!");
      setCart([]);
    },
  });

  const addToCart = (product: Product) => {
    setCart([
      ...cart,
      {
        product_id: product.id,
        requested_qty: 1,
        unit: product.unit || "pack",
      },
    ]);
  };

  const handleSubmit = () => {
    createOrderMutation.mutate({
      notes: "",
      items: cart,
    });
  };

  return (
    <div>
      {/* Search products */}
      <input
        value={search}
        onChange={(e) => setSearch(e.target.value)}
        placeholder="Search products..."
      />

      {/* Product list */}
      {products?.map((product) => (
        <ProductCard
          key={product.id}
          product={product}
          onAdd={() => addToCart(product)}
        />
      ))}

      {/* Cart */}
      <h3>Order Items ({cart.length})</h3>
      {cart.map((item, idx) => (
        <CartItem key={idx} item={item} />
      ))}

      {/* Submit */}
      <button onClick={handleSubmit}>Create Order</button>
    </div>
  );
};
