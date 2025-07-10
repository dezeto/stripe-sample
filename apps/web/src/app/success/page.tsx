"use client";

import { useEffect, useState } from "react";
import { useSearchParams } from "next/navigation";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";

export default function SuccessPage() {
	const searchParams = useSearchParams();
	const [sessionId, setSessionId] = useState<string | null>(null);

	useEffect(() => {
		const sessionId = searchParams.get("session_id");
		setSessionId(sessionId);
	}, [searchParams]);

	return (
		<div className="container mx-auto px-4 py-16">
			<div className="max-w-md mx-auto">
				<Card>
					<CardHeader className="text-center">
						<div className="mx-auto mb-4 h-12 w-12 rounded-full bg-green-100 flex items-center justify-center">
							<svg
								className="h-6 w-6 text-green-600"
								fill="none"
								stroke="currentColor"
								viewBox="0 0 24 24"
							>
								<path
									strokeLinecap="round"
									strokeLinejoin="round"
									strokeWidth={2}
									d="M5 13l4 4L19 7"
								/>
							</svg>
						</div>
						<CardTitle className="text-2xl font-bold text-green-600">
							Payment Successful!
						</CardTitle>
						<CardDescription>
							Thank you for your purchase. Your payment has been processed
							successfully.
						</CardDescription>
					</CardHeader>
					<CardContent className="text-center space-y-4">
						{sessionId && (
							<div className="text-sm text-gray-600">
								<p>
									Session ID:{" "}
									<code className="bg-gray-100 px-2 py-1 rounded text-xs">
										{sessionId}
									</code>
								</p>
							</div>
						)}
						<div className="pt-4">
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
