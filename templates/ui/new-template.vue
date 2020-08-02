<template>
  <div>
    <ValidationObserver ref="observer">
      <section slot-scope="{ validate }">
        <form @submit.prevent v-on:keyup.enter="validate().then(postValidCheck)">
          <b-field grouped group-multiline>

          //TODO:
          // for each field in *EDIT* model
          // {{.FieldRules}} ex : "min_value:1|numeric / min_value:0|numeric|required/ alpha_dash|required alpha_spaces
          // {{.FieldType}} ex: number

            <ValidationProvider rules="{{.FieldRules}}" name="{{.JSONFieldName}}">
              <b-field
                v-if="isEditMode"
                slot-scope="{ errors, valid }"
                label="{{.FieldLabel}}"
                :type="{ 'is-danger': (errors||[])[0], 'is-success': valid && validationSubmission > 0}"
                :message="errors"
              >
               <!-- // TODO:b-input is ?.number or similar needed on v-model? -->
                <b-input
                  type="{{.FieldType}}"
                  :value="null"
                  placeholder="{{.FieldPlaceHolder}}"
                  v-model="form.{{.JSONFieldName}}"
                  readonly
                ></b-input>
              </b-field>
            </ValidationProvider>

          </b-field>
          <b-button
            type="button"
            icon-left="backward"
            class="button is-small is-success"
            @click="$router.push({ name: '{{.ResourceRoute}}' }).catch(()=>{})"
            :loading="false"
          >Back</b-button>
          <b-button
            type="submit"
            :icon-left="isEditMode ? 'save':'plus'"
            class="button is-small is-success"
            @click="()=>{validate().then(postValidCheck)}"
            @keyup.enter="()=>{validate().then(postValidCheck)}"
            :loading="false"
          >{{"{{"}} isEditMode ? "Update" : "Create"{{"}}"}}</b-button>
          <b-button
            class="button is-small is-info"
            v-if="!isFormDefault"
            @click="clearForm"
            :loading="false"
          >Clear</b-button>
          <b-button
            v-if="ifReset"
            class="button is-small is-info"
            @click="resetFormToSubmitted"
            :loading="false"
          >Reset</b-button>

          <b-button
            icon-left="trash"
            class="button is-small is-danger"
            @click="onDelete{{.ModelTitleCaseName}}({id:editId, displayName: form.displayName})"
          >Delete</b-button>
        </form>
      </section>
    </ValidationObserver>
  </div>
</template>
<script>
import { parseError } from '@/util'
import {
  formMixinBuilder,
  multiDeleteMixinBuilder
} from '@/mixins'

import { mapGetters, mapActions } from 'vuex'

import isEqual from 'lodash.isequal'
import merge from 'lodash.merge'

const form = {
  {{.FormMapStatment}}
}

let formDefaults = {
  {{.FormDefaultStatement}}
  {{.JSONFieldName}}:{{.JSONDefault}}, // default null|''|undefined|false
}

let formMixin = formMixinBuilder(
  formDefaults,
  'form',
  'createOrUpdate{{.ModelTitleCaseName}}',
  undefined,
  false
)

let multiDeleteMixin = multiDeleteMixinBuilder(
  '{{.ResourceRoute}}',
  '/api/{{.ResourceRoute}}/',
  '{{.ResourceRoute}}',
  'deleteCallback',
  // `idField`,
  // `idFieldDefault`
)

export default {
  mixins: [
    formMixin,
    foldersAutoCompleteMixin,
    {{.CamelCaseModelName}}TypesAutoCompleteMixin,
    multiDeleteMixin
  ],
  beforeRouteEnter (to, from, next) {
    next()
  },
  async beforeRouteUpdate (to, from, next) {
    // TODO: if !isEditMode clear model in case they go from edit straight to new via url change. This causes the state to be the same.
    if (this.isEditMode) {
      await this.load{{.TitleCaseModelName}}(to.params.id)
    }
    next()
  },
  data () {
    return {
      validationSubmission: 0
    }
  },
  computed: {
    ...mapGetters('{{.CamelCasePluralModelName}}', {
      loading: 'isLoading'
    }),
    ifReset () {
      return this.isEditMode && !isEqual(this.form, this.submittedForm)
    },

    isEditMode () {
      return !this.$route.path.endsWith('new')
    },
    editId () {
      return this.$route.params.id || 0
    },
    idMatch () {
      return this.$route.params.id == this.form.id
    }
  },
  methods: {
    postValidCheck (isValid) {
      this.validationSubmission += 1
      if (isValid) {
        this.onSubmitForm()
      }
    },
    ...mapActions('{{.CamelCasePluralModelName}}', {
      fetch{{.TitleCaseModelName}}: 'fetchSingle',
      create{{.TitleCaseModelName}}: 'create',
      update{{.TitleCaseModelName}}: 'replace'
    }),
    get{{.TitleCaseModelName}}sDeleteMsg (objSlice) {
      objSlice = objSlice
        .slice(0, 5)
        .map(x => `<li>Id: ${x.id} - ${x.displayName}</li>`) // TODO: ID: vs Id:

      objSlice = objSlice.reduce((fin, cur) => (fin += cur), '')

      return objSlice
    },
    async onDelete{{.TitleCaseModelName}} (toDelete) {
      let toDelete{{.TitleCaseModelPluralName}}Arr = [toDelete]
      let msg = this.get{{.TitleCaseModelPluralName}}DeleteMsg(toDelete{{.TitleCaseModelName}}sArr)
      msg = this.getDeleteMsg(false, msg)
      return this.singleDelete(toDelete, msg)
    },
    deleteCallback () {
      this.$router.push({ name: '{{.ResourceRoute}}' })
    },
    async load{{.TitleCaseModelName}} (idOverride) {
      let response = {}
      try {
        let id = idOverride || this.editId
        response = await this.fetch{{.TitleCaseModelName}}({ id: id })
        // map defaults that are write only for displays
        let newDefaults = merge({}, this.defaultForm, {
          id: id,
          createdAt: response.data.createdAt,
          updatedAt: response.data.updatedAt
        })
        this.setFormDefaults(newDefaults)
        this.setSubmittedForm(response.data)
        this.validationSubmission = 0
      } catch (err) {
        if (err && err.response && err.response.status === 404) {
          this.$router.push({ name: '{{.ResourceRoute}}' }).catch(() => {})
          return
        }
        parseError(err)
      } finally {
      }
    },
    async createOrUpdate{{.TitleCaseModelName}} () {
      try {
        const config = {}
        const data = this.getForm()

        let response = {}
        if (this.isEditMode) {
          const id = this.editId
          let response = await this.update{{.TitleCaseModelName}}({
            id: id,
            data: data,
            config: config
          })
        } else {
          let response = await this.create{{.TitleCaseModelName}}({
            data: data,
            config: config
          })
        }
        this.$buefy.toast.open({
          message: this.isEditMode ? '{{.TitleCaseModelName}} Updated!' : '{{.TitleCaseModelName}} Created!',
          type: 'is-success'
        })
        this.$router.push({ name: '{{.ResourceRoute}}' }).catch(() => {})
      } catch (err) {
        parseError(err)
      }
    }
  },
  mounted () {},
  async created () {
    if (this.isEditMode) {
      await this.load{{.TitleCaseModelName}}()
    }
  }
}
</script>
