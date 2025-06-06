const { ApolloServer } = require("apollo-server");
const { ApolloGateway, IntrospectAndCompose } = require("@apollo/gateway");

// Define the port, prioritizing the environment variable
const APP_LISTEN_PORT = process.env.APP_LISTEN_PORT
  ? parseInt(process.env.APP_LISTEN_PORT, 10) // Convert to Int in base 10
  : 4000; // Default to 4000 if APP_LISTEN_PORT is not set

const gateway = new ApolloGateway({
  supergraphSdl: new IntrospectAndCompose({
    subgraphs: [
      { name: "doorcounters", url: "http://doorcounters:4001/query" },
      { name: "bms", url: "http://bms:4002/query" },
      { name: "fms", url: "http://fms:4003/query" },
      { name: "outlook", url: "http://outlook:4004/query" },
      { name: "coffee", url: "http://coffee:4005/query" },
    ],
  }),
});

const server = new ApolloServer({
  gateway,
  subscriptions: false,
});

server.listen({ host: "0.0.0.0", port: APP_LISTEN_PORT }).then(({ url }) => {
  console.log(`🚀 Server ready at ${url}`);
});
