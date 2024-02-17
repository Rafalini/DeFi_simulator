import pandas as pd
import numpy as np
import matplotlib.pyplot as plt

def generate_bitcoin_prices(start_price, drift_mean, drift_std, noise_std, num_samples):
    """
    Generates synthetic Bitcoin price data using random walk with drift and a tendency to decline.

    Parameters:
        start_price (float): Initial price of Bitcoin.
        drift_mean (float): Mean drift rate (negative for decline).
        drift_std (float): Standard deviation of drift rate.
        noise_std (float): Standard deviation of the noise.
        num_samples (int): Number of samples to generate.

    Returns:
        numpy.ndarray: Array of Bitcoin prices.
    """
    # Generate drift values with negative mean for decline
    drift_values = np.random.normal(drift_mean, drift_std, num_samples)

    # Initialize prices array
    prices = np.zeros(num_samples)
    prices[0] = start_price

    # Generate prices
    for i in range(1, num_samples):
        prices[i] = max(prices[i - 1] + drift_values[i - 1] + np.random.normal(0, noise_std), 0)  # Ensure non-negative prices

    return prices

start_price = 100  # Initial price of Bitcoin
drift_mean = 0.1   # Mean drift rate
drift_std = 0.2    # Standard deviation of drift rate
noise_std = 1     # Standard deviation of the noise
num_samples = 61   # Number of samples (60 months = 5 years)

# Generate Bitcoin prices
MKR_prices = generate_bitcoin_prices(start_price = 90, drift_mean=0.6, drift_std=0.2, noise_std=noise_std, num_samples=num_samples)
XAUT_prices = generate_bitcoin_prices(start_price = 100, drift_mean=0, drift_std=0, noise_std=10, num_samples=num_samples)
ETH_prices = generate_bitcoin_prices(start_price = 130, drift_mean=-0.5, drift_std=0.2, noise_std=noise_std, num_samples=num_samples)

# Generate the array with increasing values
time_ms = np.arange(0, num_samples * 1000, 1000)

df = pd.DataFrame({
    "time_ms": time_ms,
    "ETH": ETH_prices,
    "XAUt": XAUT_prices,
    "MKR": MKR_prices,
})

plt.plot(df["time_ms"], df['ETH'], label="ETH")
plt.plot(df["time_ms"], df['XAUt'], label="XAUt")
plt.plot(df["time_ms"], df['MKR'], label="MKR")

# Add a legend
plt.legend()

# Show the plot
plt.show()
# Write the DataFrame to a CSV file
df.to_csv('./oracle/web_app/scenario.csv', index=False)