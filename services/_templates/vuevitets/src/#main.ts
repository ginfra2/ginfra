import './assets/main.css'

import { createApp } from 'vue'
import App from './App.vue'
import router from './components/router'
import {InitApiClient, LiveApiClientInfo} from "@/api/api_manager";

const app = createApp(App)

// We need the configuration provider before anything else.  Hold up the router until config done.
const initPromise = InitApiClient()
router.beforeEach(async (to, from, next) => {
    await initPromise;
    next();
});

app.use(router)
console.log("Starting app.")
app.mount('#app')


