# DeFi_simulator

To run simulation build images & run containers:

```
docker-compose up --build --scale node=8 --scale miner=3 --remove-orphans
```

Optional scale arguments to define number of nodes or miners. Output logs will be generated into: ./src/output/ directory. Simple view of simulation state is available at localhost:5000/

Important to add value for: API-KEY-TO-COINMARKETCAP to use prices from coinmarketcap.com
