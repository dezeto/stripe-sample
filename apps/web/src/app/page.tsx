"use client";

import { useQuery } from "@tanstack/react-query";
import Image from "next/image";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";

interface Product {
	id: string;
	name: string;
	description?: string;
	images?: string[];
	metadata?: Record<string, string>;
	active: boolean;
	created: number;
	updated: number;
}

async function fetchProducts(): Promise<Product[]> {
	const response = await fetch("http://localhost:4242/list-products");
	if (!response.ok) {
		throw new Error("Failed to fetch products");
	}
	return response.json();
}

export default function Home() {
	const {
		data: products,
		isLoading,
		error,
		refetch,
	} = useQuery({
		queryKey: ["products"],
		queryFn: fetchProducts,
	});

	if (error) {
		return (
			<div className="container mx-auto px-4 py-8">
				<div className="text-center">
					<h1 className="mb-4 text-2xl font-bold text-red-600">
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
				<h1 className="mb-2 text-3xl font-bold text-gray-900 dark:text-gray-100">
					Stripe Products
				</h1>
				<p className="text-gray-600 dark:text-gray-400">
					Browse our available products
				</p>
			</div>

			{isLoading ? (
				<div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
					{Array.from({ length: 6 }, (_, i) => (
						<Card key={`skeleton-${Date.now()}-${i}`} className="w-full">
							<CardHeader>
								<Skeleton className="h-4 w-3/4" />
								<Skeleton className="h-3 w-1/2" />
							</CardHeader>
							<CardContent>
								<Skeleton className="mb-4 h-32 w-full" />
								<Skeleton className="mb-2 h-3 w-full" />
								<Skeleton className="h-3 w-2/3" />
							</CardContent>
						</Card>
					))}
				</div>
			) : (
				<div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
					{products?.map((product) => (
						<Card
							key={product.id}
							className="w-full transition-shadow hover:shadow-lg"
						>
							<CardHeader>
								<CardTitle className="text-lg font-semibold">
									{product.name}
								</CardTitle>
								{product.description && (
									<CardDescription>{product.description}</CardDescription>
								)}
							</CardHeader>
							<CardContent>
								{product.images && product.images.length > 0 && (
									<div className="mb-4">
										<Image
											src={product.images[0]}
											alt={product.name}
											width={400}
											height={128}
											className="h-32 w-full rounded-md object-cover"
										/>
									</div>
								)}
								<div className="space-y-2 text-sm text-gray-600 dark:text-gray-400">
									<p>
										<span className="font-medium">Status:</span>{" "}
										<span
											className={`rounded-full px-2 py-1 text-xs ${
												product.active
													? "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200"
													: "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200"
											}`}
										>
											{product.active ? "Active" : "Inactive"}
										</span>
									</p>
									<p>
										<span className="font-medium">Created:</span>{" "}
										{new Date(product.created * 1000).toLocaleDateString()}
									</p>
									{product.metadata &&
										Object.keys(product.metadata).length > 0 && (
											<div>
												<span className="font-medium">Metadata:</span>
												<div className="mt-1 space-y-1">
													{Object.entries(product.metadata).map(
														([key, value]) => (
															<div key={key} className="text-xs">
																<span className="rounded bg-gray-100 px-1 font-mono dark:bg-gray-800">
																	{key}:
																</span>{" "}
																{value}
															</div>
														),
													)}
												</div>
											</div>
										)}
								</div>
							</CardContent>
						</Card>
					))}
				</div>
			)}

			{products && products.length === 0 && (
				<div className="py-12 text-center">
					<h2 className="mb-2 text-xl font-semibold text-gray-900 dark:text-gray-100">
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
