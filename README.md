# blackScholes

• Developed REST API to calculate volatility surface of BTC/ETH options for a given Expiry date, Spot & Strike Price
• Implemented Black-Scholes model for options pricing to estimate volatility, leveraging Binary Search algorithm for
  optimization. Added endpoint to update volatility surface based on latest trade as input
• Utilized Deribit’s API to fetch implied volatility data for BTC/ETH options, establishing it as a reliable benchmark
  and ensuring consistency within a 2% deviation threshold as a fallback to maintain accuracy
• Integrated Rate-Limiter, Basic Authorization, Caching mechanism for the volatility surface in PostgreSQL
