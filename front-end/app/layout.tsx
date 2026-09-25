import type { Metadata } from "next";
import { DM_Sans, IBM_Plex_Mono } from "next/font/google";
import "./globals.css";

const dm = DM_Sans({
  subsets: ["latin"],
  variable: "--font-dm",
});

const plex = IBM_Plex_Mono({
  subsets: ["latin"],
  weight: ["400", "500"],
  variable: "--font-plex",
});

export const metadata: Metadata = {
  title: "Livro BTC/BRL",
  description: "Simulação de livro de ofertas em memória",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="pt-BR">
      <body className={`${dm.variable} ${plex.variable} font-sans antialiased`}>{children}</body>
    </html>
  );
}
