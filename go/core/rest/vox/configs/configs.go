// Package configs contains REST APIs for interfacing with config files.
package configs

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"disruptive/lib/configs"
	"disruptive/rest/auth"
	e "disruptive/rest/errors"
)

// GetLoadBalancers is the REST API for getting the countries config rile.
func GetLoadBalancers(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.configs.GetLoadBalancers")

	config, err := configs.Get(ctx, logCtx, "loadbalancers")
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to get countries config")
	}

	return c.JSON(http.StatusOK, config)
}

// GetLoadBalancer is the REST API for getting the a country's loadbalancer config.
func GetLoadBalancer(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.configs.GetLoadBalancer")

	country := strings.ToUpper(c.Param("country"))
	name := c.Param("name")

	config, err := configs.Get(ctx, logCtx, "loadbalancers")
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to get loadbalancers config")
	}

	lbs, ok := config[country]
	if !ok {
		return e.ErrBad(logCtx, fid, "invalid country")
	}

	l := lbs.(map[string]any)
	lb, ok := l[name].(string)
	if !ok {
		return e.ErrBad(logCtx, fid, "invalid name")
	}

	return c.JSON(http.StatusOK, map[string]string{"loadbalancer": lb})
}

// GetLocalize is the REST API for getting the a localize config.
func GetLocalize(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.configs.GetLocalize")

	version := strings.ToLower(c.Param("version"))
	language := c.Param("language")

	localize, err := configs.Get(ctx, logCtx, fmt.Sprintf("localize_%s_%s", version, language))
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to get localize config")
	}

	return c.JSON(http.StatusOK, localize)
}

// GetProducts is the REST API for getting products config file.
func GetProducts(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.configs.GetProducts")

	products, err := configs.Get(ctx, logCtx, "products")
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to get products config")
	}

	return c.JSON(http.StatusOK, products)
}

// GetProduct is the REST API for getting a product's config file.
func GetProduct(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.configs.GetProduct")

	product := c.Param("product")

	products, err := configs.Get(ctx, logCtx, "products")
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to get products config")
	}

	p, ok := products[product].(map[string]any)
	if !ok {
		return e.ErrBad(logCtx, fid, "invalid product")
	}

	return c.JSON(http.StatusOK, p)
}
