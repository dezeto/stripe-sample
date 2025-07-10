"use client";

import { useEffect, useState } from "react";

export default function MySessionsPage() {
	const [sessions, setSessions] = useState<any[]>([]);
	const [loading, setLoading] = useState(true);
	const [customerId, setCustomerId] = useState(""); // You should get this from auth/session

	useEffect(() => {
		// Replace with actual logic to get the current user's Stripe customer ID
		// const cid = window.localStorage.getItem("stripe_customer_id") || "";
		// setCustomerId(cid);

		// if (!cid) {
		// 	setLoading(false);
		// 	return;
		// }

		const customerID = "cus_SbwlgcHF9QwIIi";
		setCustomerId(customerID);

		fetch(
			`${process.env.NEXT_PUBLIC_API_URL || "http://localhost:4242"}/checkout-sessions?customer=${customerID}`,
		)
			.then((res) => res.json())
			.then((data) => setSessions(data))
			.finally(() => setLoading(false));
	}, []);

	if (loading) return <div className="p-8 text-center">Loading...</div>;

	if (!customerId) {
		return (
			<div className="p-8 text-center text-red-600">
				No customer ID found. Please log in.
			</div>
		);
	}

	return (
		<div className="container mx-auto px-4 py-8">
			<h1 className="text-2xl font-bold mb-4">Your Stripe Checkout Sessions</h1>
			{sessions.length === 0 ? (
				<div>No sessions found.</div>
			) : (
				<ul className="space-y-4">
					{sessions.map((s) => (
						<li key={s.id} className="border p-4 rounded">
							<div>
								<b>Session ID:</b> {s.id}
							</div>
							<div>
								<b>Status:</b> {s.status}
							</div>
							<div>
								<b>Amount Total:</b>{" "}
								{s.amount_total ? s.amount_total / 100 : "N/A"}{" "}
								{s.currency?.toUpperCase()}
							</div>
							<div>
								<b>Created:</b> {new Date(s.created * 1000).toLocaleString()}
							</div>
							<div>
								<b>URL:</b>{" "}
								<a
									href={s.url}
									className="text-blue-600 underline"
									target="_blank"
									rel="noopener noreferrer"
								>
									{s.url}
								</a>
							</div>
						</li>
					))}
				</ul>
			)}
		</div>
	);
}
