import { useState } from "react";
import type { ChangeEvent, FormEvent } from "react";
import PageButtons from "./components/PageButtons";
import ProductCard from "./components/Product";
import SearchBar from "./components/SearchBar";

export type Product = {
  id: string;
  vendor: string;
  brand: string;
  name: string;
  price: number;
  image: string;
  size: string;
  pricePerHundredGrams: string;
};

export type ProductResponse = {
  page: number;
  totalPages: number;
  lastPage: string;
  nextPage: string;
  count: number;
  totalItems: number;
  products: Product[];
};

export default function App() {
  const [searchInput, setSearchInput] = useState<string>("");
  const [products, setProducts] = useState<ProductResponse | null>(null);

  function postFormData(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    getNextPage(`/api/products?search=${searchInput}&page=0`);
  }

  function getNextPage(url: string) {
    fetch(encodeURI(url), {
      method: "GET",
      headers: {
        Accept: "application/json",
      },
    })
      .then((res) => res.json())
      .then((parsedResponse: ProductResponse) => setProducts(parsedResponse))
      .catch((e) => console.log(e));
  }

  function updateSearchInput(e: ChangeEvent<HTMLInputElement>) {
    setSearchInput(e.target.value);
  }

  return (
    <div className="container flex-row mx-auto">
      <SearchBar
        updateSearchInput={updateSearchInput}
        postFormData={postFormData}
      />

      {products && (
        <>
          <div className="flex p-4 flex-wrap items-center justify-evenly">
            {products.products.map((product) => (
              <ProductCard product={product} />
            ))}
          </div>
          <PageButtons
            productResponse={products}
            getNextPage={getNextPage}
            searchInput={searchInput!}
          />
        </>
      )}
    </div>
  );
}
