<template>
  <v-main>
    <v-container fluid style="height: 100vh">
      <v-row align="center" justify="center" style="height: 100%">
        <v-col cols="12" md="5" xl="3" class="text-center mt-16">
          <h2 class="text-h4 text-medium-emphasis mt-16 mb-4">
            <img
              class="mx-auto d-inline-block"
              src="@/assets/img/logo.svg"
              style="max-width: 32px"
              alt="httpSMS Logo"
            />
            Loading the httpSMS dashboard
          </h2>
          <v-progress-circular
            indeterminate
            size="160"
            class="mt-8"
            color="primary"
          />
        </v-col>
      </v-row>
    </v-container>
  </v-main>
</template>
