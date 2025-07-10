"use client";

import { useQuery } from "@tanstack/react-query";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";
import { ProductCard } from "./_components/product-card";
import type { Price } from "./_components/product-card";

async function fetchPrices(): Promise<Price[]> {
	const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:4242";
	const response = await fetch(`${apiUrl}/list-prices`);
	if (!response.ok) {
		throw new Error("Failed to fetch prices");
	}
	return response.json();
}

export default function Home() {
	const {
		data: prices,
		isLoading,
		error,
		refetch,
	} = useQuery({
		queryKey: ["prices"],
		queryFn: fetchPrices,
	});

	async function handleBuy(priceId: string) {
		const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:4242";
		const res = await fetch(`${apiUrl}/create-checkout-session`, {
			method: "POST",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({ priceId }),
		});
		if (!res.ok) {
			alert("Failed to create checkout session");
			return;
		}
		const data = await res.json();
		if (data.url) {
			window.location.href = data.url;
		} else {
			alert("No checkout URL returned");
		}
	}

	if (error) {
		return (
			<div className="container mx-auto px-4 py-8">
				<div className="text-center">
					<h1 className="text-2xl font-bold text-red-600 mb-4">
						Error Loading Products
					</h1>
					<p className="mb-4 text-gray-600">
						Failed to fetch products from the server.
					</p>
					<Button onClick={() => refetch()}>Try Again</Button>
				</div>
			</div>
		);
	}

	return (
		<div className="container mx-auto px-4 py-8">
			<div className="mb-8">
				<h1 className="text-3xl font-bold text-gray-900 dark:text-gray-100 mb-2">
					Stripe Products
				</h1>
				<p className="text-gray-600 dark:text-gray-400">
					Browse our available products
				</p>
			</div>

			{isLoading ? (
				<div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
					{Array.from({ length: 6 }, (_, i) => (
						<div key={`skeleton-${Date.now()}-${i}`} className="w-full">
							<div className="rounded-lg border bg-card text-card-foreground shadow-sm">
								<div className="flex flex-col gap-2 p-6">
									<Skeleton className="h-4 w-3/4" />
									<Skeleton className="h-3 w-1/2" />
								</div>
								<div className="p-6 pt-0">
									<Skeleton className="mb-4 h-32 w-full" />
									<Skeleton className="mb-2 h-3 w-full" />
									<Skeleton className="h-3 w-2/3" />
								</div>
							</div>
						</div>
					))}
				</div>
			) : (
				<div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
					{prices?.map((price) => (
						<ProductCard key={price.id} price={price} onBuyClick={handleBuy} />
					))}
				</div>
			)}

			{prices && prices.length === 0 && (
				<div className="py-12 text-center">
					<h2 className="text-xl font-semibold text-gray-900 dark:text-gray-100 mb-2">
						No Products Found
					</h2>
					<p className="text-gray-600 dark:text-gray-400">
						There are no products available at the moment.
					</p>
				</div>
			)}
		</div>
	);
}
