
<template>
  <div v-if="LiveApiClientInfo().serverFault">
      <h2>Service unavailable.</h2>
      <pre>{{ LiveApiClientInfo().serverFault }}</pre>
  </div>
  <div v-else>
    <div v-if="serviceerror">
      <h2>Service Error</h2>
      <p>Service is inoperable.</p>
      <pre>{{ serviceerror }}</pre>
    </div>

      <div v-else>
        Stuff here.
      </div>

  </div>
</template>

<script lang="ts">
import {type userstateinfo, LiveApiClientInfo} from "@/api/api_manager"
import App from "@/App.vue"
export default {
  components: {App},
  data() {
    return {
      reqerror: "",
      serviceerror: "",
    };
  },
  methods: {
    LiveApiClientInfo() {
      return LiveApiClientInfo
    }
  },
}
</script>
<style>
form {
  padding: 5px;
  border: 1px solid black;
  border-radius: 3px;
}

input {
  padding: 4px 8px;
  margin: 4px;
}

span {
  font-size: 18px;
  margin: 4px;
  font-weight: 500;
}

.doit {
  font-size: 16px
}

.inputusername {
  font-weight: 200;
}

.results {
  font-weight: 150;
}

</style>