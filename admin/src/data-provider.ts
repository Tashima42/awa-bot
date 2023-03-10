import {
    CreateParams,
    DataProvider,
    DeleteManyParams,
    DeleteParams,
    GetListParams,
    GetManyParams,
    GetManyReferenceParams, GetOneParams, UpdateManyParams, UpdateParams,
    fetchUtils,
} from "react-admin";
import { stringify } from "query-string";

const httpClient = fetchUtils.fetchJson

function getDataProvider(apiUrl: string): DataProvider {
    return {
        getList(resource: string, params: GetListParams): Promise<any> {
            const { page, perPage } = params.pagination;
            const query = {
                limit: perPage,
                skip: (page - 1) * perPage,
            }
            const url = `${apiUrl}/${resource}?${stringify(query)}`;
            let apikeyString = ""
            let userIDString = ""
            const apikey = localStorage.getItem('apikey')
            const userID = localStorage.getItem('userID')
            if (apikey) {
                console.log("apikey", apikey)
                apikeyString = apikey
            }
            if (userID) {
                console.log("userID", userID)
                userIDString = userID
            }
            return httpClient(url, {
                headers: new Headers({
                    'Content-Type': 'application/json',
                    'x-apikey': apikeyString,
                    'x-user-id': userIDString,
                }),
            }).then(({ headers, json }) => ({
                data: json.waters,
                total: json.total,
            }));
        },
        create(resource: string, params: CreateParams): Promise<any> {
            return Promise.resolve(undefined);
        },
        delete(resource: string, params: DeleteParams): Promise<any> {
            return Promise.resolve(undefined);
        },
        deleteMany(resource: string, params: DeleteManyParams): Promise<any> {
            return Promise.resolve(undefined);
        },
        getMany(resource: string, params: GetManyParams): Promise<any> {
            return Promise.resolve(undefined);
        },
        getManyReference(resource: string, params: GetManyReferenceParams): Promise<any> {
            return Promise.resolve(undefined);
        },
        getOne(resource: string, params: GetOneParams): Promise<any> {
            return Promise.resolve(undefined);
        },
        update(resource: string, params: UpdateParams): Promise<any> {
            return Promise.resolve(undefined);
        },
        updateMany(resource: string, params: UpdateManyParams): Promise<any> {
            return Promise.resolve(undefined);
        }
    }
}

export default getDataProvider;