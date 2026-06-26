package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleBuscarCEP(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cep, _ :=getString(args, "cep")
	u := fmt.Sprintf("https://viacep.com.br/ws/%s/json/", url.QueryEscape(cep))
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("erro ao consultar CEP: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("erro ao ler resposta: " + e.Error())
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("erro ao decodificar JSON: " + e.Error())
}

	if data["erro"] != nil {
		return err("CEP não encontrado")
}

	return ok(fmt.Sprintf("CEP %s: %s, %s - %s/%s", cep, data["logradouro"], data["bairro"], data["localidade"], data["uf"]))
}

func HandleConsultarCNPJ(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cnpj, _ :=getString(args, "cnpj")
	u := fmt.Sprintf("https://www.receitaws.com.br/v1/cnpj/%s", url.QueryEscape(cnpj))
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("erro ao consultar CNPJ: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("erro ao ler resposta: " + e.Error())
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("erro ao decodificar JSON: " + e.Error())
}

	if data["status"] == "ERROR" {
		return err(data["message"].(string))
}

	nome, _ := data["nome"].(string)
	return ok(fmt.Sprintf("CNPJ %s: %s - Situação: %s", cnpj, nome, data["situacao"]))
}