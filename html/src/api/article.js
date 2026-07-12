import request from "../utils/request";
import { createCRUDApi, extendApi } from "../utils/apiFactory";
import { normalizeListResponse } from "../utils/normalize";

const baseArticleApi = createCRUDApi("articles");

const articleApi = baseArticleApi;

export async function getArticleList(params) {
  const res = await articleApi.list(params);
  return normalizeListResponse(res);
}

export function getArticleDetail(id) {
  return articleApi.detail(id);
}

export function createArticle(data) {
  return articleApi.create(data);
}

export function updateArticle(id, data) {
  return articleApi.update(id, data);
}

export function deleteArticle(id) {
  return articleApi.delete(id);
}
