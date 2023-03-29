<script setup lang="ts">
defineProps<{}>()
</script>

<template>
  <div class="greetings">
    <h3>
      Hello
    </h3>
  </div>
</template>

<script lang="ts">
import initSqlJs from "sql.js";
import sqlWasmUrl from "sql.js/dist/sql-wasm.wasm?url";

console.log(sqlWasmUrl)
const sqljs = await initSqlJs({locateFile: () => sqlWasmUrl});
console.log(sqljs)
const dbFile = await fetch("db.sqlite3");
console.log(dbFile);
const dbBuf = await dbFile.arrayBuffer()
console.log(dbBuf)
const db = new sqljs.Database(new Uint8Array(dbBuf));
console.log(db)

const result = db.exec("SELECT * FROM trains ORDER BY start_ts DESC LIMIT 1;")
console.log(result)

</script>


<style scoped>
h1 {
  font-weight: 500;
  font-size: 2.6rem;
  top: -10px;
}

h3 {
  font-size: 1.2rem;
}

.greetings h1,
.greetings h3 {
  text-align: center;
}

@media (min-width: 1024px) {
  .greetings h1,
  .greetings h3 {
    text-align: left;
  }
}
</style>
