import {AuthProvider, fetchUtils} from "react-admin";

const httpClient = fetchUtils.fetchJson;

function getAuthProvider(apiUrl: string): AuthProvider {
    return {
        login(params: any): Promise<void> {
            const { username, password } = params;
            const url = `${apiUrl}/whoami`;
            return httpClient(url, {
                method: 'GET',
                headers: new Headers({
                    'Content-Type': 'application/json',
                    'x-user-id': username,
                    'x-apikey': password,
                }),
            }).then(({ status}) => {
                if (status !== 200) {
                    return Promise.reject();
                }
                localStorage.setItem('userID', username);
                localStorage.setItem('apikey', password);
            })
        },
        checkAuth(): Promise<void> {
            const userID = localStorage.getItem('userID');
            if (userID) {
                return Promise.resolve();
            }
            return Promise.reject();
        },
        checkError(error: any): Promise<void> {
            console.log("login error:", error)
            return Promise.resolve(undefined);
        },
        getIdentity(): Promise<any> {
            const userInfo = { id: localStorage.getItem('userID'), }
            if (userInfo.id) {
                return Promise.resolve(userInfo);
            }
            return Promise.resolve();
        },
        getPermissions(): Promise<any> {
            return Promise.resolve('');
        },
        logout(): Promise<void | false | string> {
            localStorage.removeItem('userID');
            localStorage.removeItem('apikey');
            return Promise.resolve();
        }
    }
}

export default getAuthProvider