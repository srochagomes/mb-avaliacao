const messages: Record<string, string> = {
  "insufficient balance": "Saldo disponível insuficiente.",
  "invalid amount": "Valor inválido. Use um número maior que zero, com até 8 casas.",
  "invalid asset": "Ativo inválido. Use BTC ou BRL.",
  "invalid side": "Lado inválido. Use compra ou venda.",
  "invalid instrument": "Instrumento inválido. Este livro é BTC-BRL.",
  "invalid json": "Não foi possível ler o pedido.",
  "label required": "Dê um nome à carteira.",
  "label too long": "O nome da carteira pode ter no máximo 40 caracteres.",
  "account_id required": "Escolha uma carteira.",
  "order id required": "Ordem sem identificador.",
  "account not found": "Carteira não encontrada.",
  "order not found": "Ordem não encontrada.",
  "order not open": "Só é possível cancelar uma ordem aberta.",
  "api unavailable": "A API não respondeu. Confira se o serviço está no ar.",
};

export function explain(error: string) {
  return messages[error] ?? error;
}

export function formatAmount(raw: string) {
  if (!raw) return "0";
  const negative = raw.startsWith("-");
  const [whole, frac] = raw.replace("-", "").split(".");
  const grouped = whole.replace(/\B(?=(\d{3})+(?!\d))/g, ".");
  const body = frac ? `${grouped},${frac}` : grouped;
  return negative ? `-${body}` : body;
}

export function shortId(id: string) {
  return id.slice(0, 8);
}
