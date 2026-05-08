import type { Metadata } from "next";
import { Inter } from "next/font/google";
import { AuthStoreHydrator } from "@/components/auth/AuthStoreHydrator";
import { Footer } from "@/components/Footer";
import { Header } from "@/components/Header";
import { ThemeProvider } from "@/components/providers/ThemeProvider";
import { getServerAuthSnapshot } from "@/lib/server-auth";
import "./globals.css";

const inter = Inter({
  variable: "--font-inter",
  subsets: ["latin", "cyrillic"],
});

export const metadata: Metadata = {
  title: "Беру Смену . ру — срочные смены для ПВЗ",
  description:
    "Беру Смену . ру соединяет владельцев ПВЗ и исполнителей для срочных смен и подмен.",
};

export default async function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  const authSnapshot = await getServerAuthSnapshot();

  return (
    <html lang="ru" suppressHydrationWarning>
      <body className={`${inter.variable} flex min-h-screen flex-col font-sans antialiased`}>
        <ThemeProvider>
          <AuthStoreHydrator snapshot={authSnapshot}>
            <Header />
            <main className="min-h-0 w-full flex-1 pt-[var(--header-height)]">
              {children}
            </main>
            <Footer />
          </AuthStoreHydrator>
        </ThemeProvider>
      </body>
    </html>
  );
}
