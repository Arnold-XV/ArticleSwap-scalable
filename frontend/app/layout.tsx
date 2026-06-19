import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "ArticleSwap",
  description: "Platform pertukaran artikel scalable dan resilien"
};

export default function RootLayout({
  children
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="id">
      <body>{children}</body>
    </html>
  );
}

