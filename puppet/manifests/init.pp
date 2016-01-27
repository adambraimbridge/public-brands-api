class public_brands_api {

  $configParameters = hiera('configParameters','')

  class { "go_service_profile" :
    service_module => $module_name,
    service_name => 'public-brands-api',
    configParameters => $configParameters
  }

}
