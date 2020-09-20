import createCrudModule from 'vuex-crud'
import client from '@/axios'
import {
  parseList
} from '@/util'

export default createCrudModule({
  resource: '{{.Resource}}',
  client: client,
  parseError: function (res) {
    return res
  },
  parseList
})
