import {Admin, Resource} from "react-admin";
import {WaterList} from "./water-list";
import getDataProvider from "./data-provider";
import getAuthProvider from "./auth-provider";

const apiUrl = "http://127.0.0.1:8096"
const authProvider = getAuthProvider(apiUrl);
const dataProvider = getDataProvider(apiUrl);

const App = () => (
    <Admin authProvider={authProvider} dataProvider={dataProvider} requireAuth>
        <Resource name="water" list={WaterList} />
    </Admin>
)
export default App;