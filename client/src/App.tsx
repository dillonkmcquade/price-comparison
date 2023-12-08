import { useState } from "react";
import type { ChangeEvent, FormEvent } from "react";
import PageButtons from "./components/PageButtons";
import ProductCard from "./components/Product";

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
    getNextPage(
      `/api/products?search=${searchInput}&page=0`,
    );
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
      <form onSubmit={postFormData}>
        <label
          htmlFor="default-search"
          className="mb-2 text-sm font-medium text-gray-900 sr-only dark:text-gray-300"
        >
          Search
        </label>
        <div className="relative">
          <div className="flex absolute inset-y-0 left-0 items-center pl-3 pointer-events-none">
            <svg
              className="w-5 h-5 text-gray-500 dark:text-gray-400"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
              xmlns="http://www.w3.org/2000/svg"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
              ></path>
            </svg>
          </div>
          <input
            type="search"
            id="default-search"
            className="block p-4 pl-10 w-full text-sm text-gray-900 bg-gray-50 rounded-lg border border-gray-300 focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500"
            placeholder="Search products..."
            onChange={updateSearchInput}
            required
          ></input>
          <button
            type="submit"
            className="text-white absolute right-2.5 bottom-2.5 bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:outline-none focus:ring-blue-300 font-medium rounded-lg text-sm px-4 py-2 dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800"
          >
            Search
          </button>
        </div>
      </form>
      <div>
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
    </div>
  );
}
