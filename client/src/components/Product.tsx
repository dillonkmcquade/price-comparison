import { Product } from "../App";

export default function ProductCard({ product }: { product: Product }) {
  return (
    <div className="w-[250px] gap-2" key={product.id}>
      <img className="w-30 h-30" src={product.image} loading="lazy" />
      <div className="font-bold">{product.name}</div>
      <div>{product.brand}</div>
      <div>{product.vendor}</div>
      <div>{product.size}</div>
      <div className="font-bold">${product.price}</div>
      <div className="text-gray-500 italic">{product.pricePerHundredGrams}</div>
    </div>
  );
}
