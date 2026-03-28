import { createJiti } from "../../../../../node_modules/.pnpm/jiti@2.6.1/node_modules/jiti/lib/jiti.mjs";

const jiti = createJiti(import.meta.url, {
  "interopDefault": true,
  "alias": {
    "@vben-core/design": "/Users/linshen/Cursor/Full-stack-TestV1/vue-vben-admin/packages/@core/base/design"
  },
  "transformOptions": {
    "babel": {
      "plugins": []
    }
  }
})

/** @type {import("/Users/linshen/Cursor/Full-stack-TestV1/vue-vben-admin/packages/@core/base/design/src/index.js")} */
const _module = await jiti.import("/Users/linshen/Cursor/Full-stack-TestV1/vue-vben-admin/packages/@core/base/design/src/index.ts");

export default _module?.default ?? _module;