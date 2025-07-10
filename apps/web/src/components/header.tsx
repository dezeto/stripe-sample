"use client";
import Link from "next/link";
import { ModeToggle } from "./mode-toggle";
import {
	NavigationMenu,
	NavigationMenuList,
	NavigationMenuItem,
	NavigationMenuLink,
} from "@/components/ui/navigation-menu";

export default function Header() {
	return (
		<div>
			<div className="flex flex-row items-center justify-between px-2 py-1">
				<NavigationMenu>
					<NavigationMenuList>
						<NavigationMenuItem>
							<NavigationMenuLink asChild>
								<Link href="/">Home</Link>
							</NavigationMenuLink>
						</NavigationMenuItem>
						<NavigationMenuItem>
							<NavigationMenuLink asChild>
								<Link href="/my-sessions">My Sessions</Link>
							</NavigationMenuLink>
						</NavigationMenuItem>
					</NavigationMenuList>
				</NavigationMenu>
				<div className="flex items-center gap-2">
					<ModeToggle />
				</div>
			</div>
			<hr />
		</div>
	);
}
