const { ApolloServer } = require("apollo-server");
const { ApolloGateway, IntrospectAndCompose } = require("@apollo/gateway");

const gateway = new ApolloGateway({
  supergraphSdl: new IntrospectAndCompose({
    subgraphs: [
      { name: "doorcounters", url: "http://localhost:4001/query" },
      { name: "bms", url: "http://localhost:4002/query" },
      { name: "fms", url: "http://localhost:4003/query" },
      { name: "coffee", url: "http://localhost:4005/query" },
    ],
  }),
});

const server = new ApolloServer({
  gateway,
  subscriptions: false,
});

server.listen().then(({ url }) => {
  console.log(`🚀 Server ready at ${url}`);
});
