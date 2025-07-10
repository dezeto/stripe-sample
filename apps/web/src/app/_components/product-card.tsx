import Image from "next/image";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";

export interface Product {
	id: string;
	name: string;
	description?: string;
	images?: string[];
	metadata?: Record<string, string>;
	active: boolean;
	created: number;
	updated: number;
}

export interface Price {
	id: string;
	object: string;
	active: boolean;
	billing_scheme: string;
	created: number;
	currency: string;
	custom_unit_amount: null;
	livemode: boolean;
	lookup_key: null;
	metadata: Record<string, string>;
	nickname: null;
	product: Product;
	recurring: null;
	tax_behavior: string;
	tiers_mode: null;
	transform_quantity: null;
	type: string;
	unit_amount: number;
	unit_amount_decimal: string;
}

interface ProductCardProps {
	price: Price;
	onBuyClick?: (priceId: string) => void;
}

function formatPrice(amount: number, currency: string): string {
	return new Intl.NumberFormat("en-US", {
		style: "currency",
		currency: currency.toUpperCase(),
	}).format(amount / 100);
}

export function ProductCard({ price, onBuyClick }: ProductCardProps) {
	const product = price.product;

	return (
		<Card key={price.id} className="w-full transition-shadow hover:shadow-lg">
			<CardHeader>
				<CardTitle className="font-semibold text-lg">{product.name}</CardTitle>
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

				<div className="mb-4">
					<div className="text-2xl font-bold text-gray-900 dark:text-gray-100">
						{formatPrice(price.unit_amount, price.currency)}
					</div>
					<div className="text-sm text-gray-500 dark:text-gray-400">
						{price.type === "one_time" ? "One-time payment" : "Recurring"}
					</div>
				</div>

				{onBuyClick && (
					<button
						type="button"
						className="w-full mb-4 rounded bg-blue-600 px-4 py-2 font-semibold text-white transition-colors hover:bg-blue-700"
						onClick={() => onBuyClick(price.id)}
					>
						Buy Now
					</button>
				)}

				<div className="space-y-2 text-sm text-gray-600 dark:text-gray-400">
					<p>
						<span className="font-medium">Status:</span>{" "}
						<span
							className={`rounded-full px-2 py-1 text-xs ${
								price.active
									? "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200"
									: "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200"
							}`}
						>
							{price.active ? "Active" : "Inactive"}
						</span>
					</p>
					<p>
						<span className="font-medium">Created:</span>{" "}
						{new Date(price.created * 1000).toLocaleDateString()}
					</p>
					{product.metadata && Object.keys(product.metadata).length > 0 && (
						<div>
							<span className="font-medium">Metadata:</span>
							<div className="mt-1 space-y-1">
								{Object.entries(product.metadata).map(([key, value]) => (
									<div key={key} className="text-xs">
										<span className="rounded bg-gray-100 px-1 font-mono dark:bg-gray-800">
											{key}:
										</span>{" "}
										{value}
									</div>
								))}
							</div>
						</div>
					)}
				</div>
			</CardContent>
		</Card>
	);
}
