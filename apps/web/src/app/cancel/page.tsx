"use client";

import Link from "next/link";
import { Button } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";

export default function CancelPage() {
	return (
		<div className="container mx-auto px-4 py-16">
			<div className="max-w-md mx-auto">
				<Card>
					<CardHeader className="text-center">
						<div className="mx-auto mb-4 h-12 w-12 rounded-full bg-red-100 flex items-center justify-center">
							<svg
								className="h-6 w-6 text-red-600"
								fill="none"
								stroke="currentColor"
								viewBox="0 0 24 24"
							>
								<path
									strokeLinecap="round"
									strokeLinejoin="round"
									strokeWidth={2}
									d="M6 18L18 6M6 6l12 12"
								/>
							</svg>
						</div>
						<CardTitle className="text-2xl font-bold text-red-600">
							Payment Canceled
						</CardTitle>
						<CardDescription>
							Your payment was canceled. No charges have been made to your
							account.
						</CardDescription>
					</CardHeader>
					<CardContent className="text-center space-y-4">
						<p className="text-gray-600">
							You can continue shopping or try again when you're ready.
						</p>
						<div className="pt-4 space-y-2">
							<Button asChild className="w-full">
								<Link href="/">Continue Shopping</Link>
							</Button>
						</div>
					</CardContent>
				</Card>
			</div>
		</div>
	);
}
