// configuração de runtime da aplicação.
// em produção, o entrypoint do container deve gerar este arquivo a partir
// da variável de ambiente API_URL — assim a mesma imagem serve em qualquer
// ambiente. este conteúdo é apenas o default para desenvolvimento local.
window.ENV = {
  API_URL: "http://localhost:8080",
};
