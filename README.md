#blackScholes
Derived volatility surface of BTC and ETH options based on given expiry dates, spot prices, and strike prices

#About
Integrated the Black-Scholes model for options pricing, employing a Binary Search algorithm to optimize the estimation of volatility.
Created an endpoint to update the volatility surface dynamically, utilizing the latest trade data as input. Leveraged Deribit’s API to retrieve implied volatility data for BTC/ETH options, ensuring reliable benchmarks and maintaining consistency within a 2% deviation threshold to ensure accuracy.
