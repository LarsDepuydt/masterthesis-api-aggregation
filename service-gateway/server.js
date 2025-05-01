const { ApolloServer } = require("apollo-server");
const { ApolloGateway, IntrospectAndCompose } = require("@apollo/gateway");

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

server.listen().then(({ url }) => {
  console.log(`🚀 Server ready at ${url}`);
});
