<template>
  <div>
    <ValidationObserver ref="observer">
      <section slot-scope="{ validate }">
        <form @submit.prevent v-on:keyup.enter="validate().then(postValidCheck)">
          <b-field grouped group-multiline>

            // TODO: same field loop
            {{range .ModelFieldsMeta}}
            <ValidationProvider rules="{{.FieldRule}}" name="{{.FieldName}}">
              <b-field
                slot-scope="{ errors, valid }"
                label="{{.FieldLabel}}"
                :type="{ 'is-danger': errors[0], 'is-success': valid && isSearchSubmitted}"
                :message="errors"
              >
              <!-- TODO: is .number or similar needed on v-model?  -->
                <b-input type="{{.FieldType}}" :value="null" placeholder="{{.FieldLabel}}" v-model="search.{{.FieldName}}"></b-input>
              </b-field>
            </ValidationProvider>
            {{end}}
            <b-field label="Per Page">
              <b-select v-model.number="pageInfoReadOnlyPerPage" @input="onPerPageChange">
                <option value="1">1</option>
                <option value="5">5</option>
                <option value="10">10</option>
                <option value="15">15</option>
                <option value="20">20</option>
              </b-select>
            </b-field>
            <b-button
              icon-left="plus"
              class="button is-small is-success"
              @click="$router.push({name:'new{{.ModelTitleName}}'}).catch(()=>{})"
              :loading="false"
            >New</b-button>
            <b-button
              icon-left="sync"
              class="button is-small is-success"
              @click="fetchListData"
              :loading="false"
            >Refresh</b-button>
            <b-button
              type="submit"
              icon-left="search"
              class="button is-small is-success"
              @click="()=>{validate().then(postValidCheck)}"
              @keyup.enter="()=>{validate().then(postValidCheck)}"
              :loading="false"
            >Search</b-button>
            <b-button
              v-if="!isSearchDefault"
              class="button is-small is-info"
              @click="clearSearch"
              :loading="false"
            >Clear</b-button>
            <b-button
              v-if="isSearchSubmitted"
              class="button is-small is-info"
              @click="resetSearch"
              :loading="false"
            >Reset</b-button>
          </b-field>
        </form>
      </section>
    </ValidationObserver>

    <b-table
      ref="zeTable"
      :loading="loading"
      :data="{{.CamelCaseModelName}}TableData"
      :striped="true"
      :hoverable="true"
      :mobile-cards="true"
      checkable
      backend-pagination
      backend-sorting
      :checked-rows.sync="toDelete{{.ModelNamePluralTitleCase}}Arr"
      :is-row-checkable="(row) => true"
      :custom-is-checked="checkedHandler"
      :default-sort-direction="pageInfo.sortDir"
      :default-sort="pageInfo.sortBy"
      :checkbox-position="'left'"
      :paginated="true"
      :pagination-simple="false"
      :per-page="pageInfoReadOnlyPerPage"
      :current-page="pageInfo.page"
      :total="pageInfo.total"
      @sort="onSort"
      @page-change="onPageChange"
      :pagination-position="'bottom'"
      aria-next-label="Next page"
      aria-previous-label="Previous page"
      aria-page-label="Page"
      aria-current-label="Current page"
    >
      <template slot-scope="props">
        // FOR EACH FIELD ADD COLUMN
        // FOR NUMBER add numeric, sortable
        // sortable
        // TODO:

        // TODO: if ID
        <b-table-column field="id" label="ID" width="40" numeric sortable>
          <router-link
            :to="{ name: 'edit{{.TitleCaseModelName}}', params: { id:props.row.id, model: props.row {{"}}"}}"
          >{{"{{"}} props.row.id{{"}}"}}</router-link>
        </b-table-column>
        {{range .ModelFieldsMeta}}
        <b-table-column
          field="{{.FieldName}}"
          label="{{.FieldLabel}}"
          {{.ColModifers}}
        >{{"{{"}} props.row.{{.FieldName}} {{"}}"}}</b-table-column>
        {{end}}
        <b-table-column label="Edit">
          <router-link :to="{ name: 'edit{{.TitleCaseModelName}}', params: { id:props.row.id, model: props.row {{"}}"}}">
            <b-icon icon="edit" size="is-small"></b-icon>
          </router-link>

          <b-icon icon="minus-circle" size="is-small" @click.native="onDelete{{.TitleCaseModelName}}(props.row)"></b-icon>
        </b-table-column>
      </template>

      <template slot="empty">
        <section class="section">
          <div class="content has-text-grey has-text-centered">
            <p>
              <b-icon icon="frown" size="is-medium"></b-icon>
            </p>
            <p>No {{.ModelNamePluralTitleCase}}.</p>
          </div>
        </section>
      </template>

      <template slot="footer"></template>
      <template slot="bottom-left">
        <b>Total checked</b>
        : {{"{{"}}toDelete{{.ModelNamePluralTitleCase}}Arr.length {{"}}"}}
        <b-button
          icon-left="trash"
          v-if="toDelete{{.ModelNamePluralTitleCase}}Arr.length >0"
          class="button is-small is-danger"
          @click="confirmMultiDelete()"
        >Delete Checked</b-button>
        <b-button
          class="button is-small is-info"
          v-if="toDelete{{.ModelNamePluralTitleCase}}Arr.length >0"
          @click="clearToDelete{{.ModelNamePluralTitleCase}}Arr"
        >Clear Checked</b-button>
      </template>
    </b-table>
  </div>
</template>

<script>
import { parseError } from "@/util";
import {
  formMixinBuilder,
  pageMixinBuilder,
  multiDeleteMixinBuilder,
} from "@/mixins";
import { mapGetters, mapActions } from "vuex";

import isEqual from "lodash.isequal";
import merge from "lodash.merge";

let pageInfoDefaults = {
  perPage: 20,
  page: 1,
  sortBy: "id",
  sortDir: "asc",
  total: 0
};

const colOverride = {
  {{.COLOverrideStatement}}
};
let pgMixin = pageMixinBuilder(pageInfoDefaults, "fetchListData", colOverride);

let multiDeleteMixin = multiDeleteMixinBuilder(
  '{{.ResourceRoute}}',
  '/api/{{.ResourceRoute}}/',
  '{{.ResourceRoute}}',
  'fetchListData',
  // `idField`,
  // `idFieldDefault`
)

let formDefaults = {
  {{.FormDefaultStatement}}
}

let formMixin = formMixinBuilder(formDefaults, "search", "fetchListData");

let routeDefaults = {
  query: {
    ...pageInfoDefaults
  }
};
routeDefaults.query.total = undefined;

export default {
  mixins: [
    formMixin,
    pgMixin,
    multiDeleteMixin,
  ],
  beforeRouteEnter(to, from, next) {
    let nRoute = merge({}, routeDefaults, to);
    if (!isEqual(nRoute.query, to.query)) {
      next(nRoute);
      return;
    }
    next();
  },
  async beforeRouteUpdate(to, from, next) {
    let nRoute = merge({}, routeDefaults, to);
    if (!isEqual(nRoute, to)) {
      next(nRoute);
      return;
    }
    this.loadPageStateFromQuery(to.query);
    if (!to.params.skipLoad && !isEqual(to.query, from.query)) {
      await this.fetchListData();
    }
    next();
  },
  data() {
    return {
      validationSubmission: 0
    };
  },
  computed: {
    ...mapGetters("{{.CamelCasePlural}}", {
      {{.CamelCasePlural}}: "list",
      loading: "isLoading"
    }),
    {{.CamelCaseModelName}}TableData() {
      return this.{{.CamelCasePlural}}.map(x => {
        return { ...x };
      });
    },
    deleteMsgObjectsLabel() {
      let objStr = this.toDelete{{.ModelNamePluralTitleCase}}Arr.map(x => merge({}, x));
      return this.get{{.ModelNamePluralTitleCase}}DeleteMsg(objStr);
    },
    paramsSearch() {
      let filter = {
        $or: []
      };
      {{.SearchStatement}}


      if (!this.isSearchSubmitted || (filter && filter.length === 0)) {
        filter = undefined;
      }

      filter = JSON.stringify(filter);

      return {
        filter: filter
      };
    },
    params() {
      let p = merge({}, this.paramsPageInfo, this.paramsSearch);
      return p;
    }
  },
  methods: {
    get{{.ModelNamePluralTitleCase}}DeleteMsg(objSlice) {
      objSlice = objSlice
        .slice(0, 5)
        .map(x => `<li>Id: ${x.id} - ${x.displayName}</li>`); // TODO: ID: vs Id:

      objSlice = objSlice.reduce((fin, cur) => (fin += cur), "");

      return objSlice;
    },
    async onDelete{{.TitleCaseModelName}}(toDelete) {
      let toDelete{{.ModelNamePluralTitleCase}}Arr = [toDelete];
      let msg = this.get{{.ModelNamePluralTitleCase}}DeleteMsg(toDelete{{.ModelNamePluralTitleCase}}Arr);
      msg = this.getDeleteMsg(false, msg);
      return this.singleDelete(toDelete, msg);
    },
    postValidCheck(isValid) {
      this.validationSubmission += 1;
      if (isValid) {
        this.onSubmitSearch();
      }
    },
    loadPageStateFromQuery(query) {
      this.setSearchFromQuery(query);
      this.setPageInfoFromQuery(query, true, undefined);
    },
    ...mapActions("{{.CamelCasePlural}}", {
      fetch{{.ModelNamePluralTitleCase}}: "fetchList"
    }),
    async fetchListData() {
      try {
        const config = {
          params: this.params
        };
        let response = await this.fetch{{.ModelNamePluralTitleCase}}({
          config: config
        });
        let pageInfo = this.pageToPageInfo(response.page);
        this.setPageQuery(pageInfo);
        this.setPageInfo(pageInfo);
        // runs table sort to update ui displayed sorted column after
        await this.$nextTick(async () => {
          this.isPageEventsEnabled = false;
          await this.$refs.zeTable.initSort();
          this.isPageEventsEnabled = true;
        });
      } catch (err) {
        parseError(err);
      }
    }
  },
  mounted() {},
  async created() {
    this.loadPageStateFromQuery(this.$route.query);
    await this.fetchListData();
  }
};
</script>
