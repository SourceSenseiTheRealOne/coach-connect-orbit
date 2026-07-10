export interface LocalServiceOrigins {
  goApi: string;
  social: string;
  marketplace: string;
}

export const localServiceOrigins: Readonly<LocalServiceOrigins> = Object.freeze(
  {
    goApi: "http://127.0.0.1:9000",
    social: "http://127.0.0.1:3001",
    marketplace: "http://127.0.0.1:3002",
  },
);
