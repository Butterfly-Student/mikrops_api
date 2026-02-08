package gin_inbound_adapter

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/palantir/stacktrace"

	"mikrops/internal/domain"
	"mikrops/internal/model"
	inbound_port "mikrops/internal/port/inbound"
	outbound_port "mikrops/internal/port/outbound"
	"mikrops/utils/activity"
	"mikrops/utils/crypto"
)

type mikrotikAdapter struct {
	domain   domain.Domain
	httpPort outbound_port.HttpPort
}

func NewMikrotikAdapter(domain domain.Domain, httpPort outbound_port.HttpPort) inbound_port.MikrotikHttpPort {
	return &mikrotikAdapter{domain: domain, httpPort: httpPort}
}

func (h *mikrotikAdapter) getNasFromParam(c *gin.Context) (model.Nas, string, error) {
	nasID := c.Param("id")
	ctx := activity.NewContext("mikrotik_get_nas")
	nas, err := h.domain.Nas().FindByID(ctx, nasID)
	if err != nil {
		return model.Nas{}, "", stacktrace.Propagate(err, "failed to find NAS")
	}
	password, err := crypto.Decrypt(nas.PasswordEncrypted)
	if err != nil {
		return model.Nas{}, "", stacktrace.Propagate(err, "failed to decrypt NAS password")
	}
	return nas, password, nil
}

func (h *mikrotikAdapter) mikrotikPort() outbound_port.MikrotikPort {
	return h.httpPort.Mikrotik()
}

// PPPoE Secrets

func (h *mikrotikAdapter) ListPPPoESecrets(a any) error {
	c := a.(*gin.Context)

	nas, password, err := h.getNasFromParam(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	results, err := h.mikrotikPort().ListPPPoESecrets(nas.Host, nas.RestPort, nas.Username, password, nas.UseSSL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: results})
	return nil
}

func (h *mikrotikAdapter) CreatePPPoESecret(a any) error {
	c := a.(*gin.Context)

	nas, password, err := h.getNasFromParam(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	var payload model.MikrotikPPPoESecretInput
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return nil
	}

	err = h.mikrotikPort().CreatePPPoESecret(nas.Host, nas.RestPort, nas.Username, password, nas.UseSSL, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusCreated, model.Response{Success: true})
	return nil
}

func (h *mikrotikAdapter) UpdatePPPoESecret(a any) error {
	c := a.(*gin.Context)

	nas, password, err := h.getNasFromParam(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	sid := c.Param("sid")

	var payload model.MikrotikPPPoESecretInput
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return nil
	}

	err = h.mikrotikPort().UpdatePPPoESecret(nas.Host, nas.RestPort, nas.Username, password, nas.UseSSL, sid, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true})
	return nil
}

func (h *mikrotikAdapter) DeletePPPoESecret(a any) error {
	c := a.(*gin.Context)

	nas, password, err := h.getNasFromParam(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	sid := c.Param("sid")

	err = h.mikrotikPort().DeletePPPoESecret(nas.Host, nas.RestPort, nas.Username, password, nas.UseSSL, sid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true})
	return nil
}

func (h *mikrotikAdapter) DisablePPPoESecret(a any) error {
	c := a.(*gin.Context)

	nas, password, err := h.getNasFromParam(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	sid := c.Param("sid")

	err = h.mikrotikPort().DisablePPPoESecret(nas.Host, nas.RestPort, nas.Username, password, nas.UseSSL, sid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true})
	return nil
}

func (h *mikrotikAdapter) EnablePPPoESecret(a any) error {
	c := a.(*gin.Context)

	nas, password, err := h.getNasFromParam(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	sid := c.Param("sid")

	err = h.mikrotikPort().EnablePPPoESecret(nas.Host, nas.RestPort, nas.Username, password, nas.UseSSL, sid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true})
	return nil
}

// PPPoE Profiles

func (h *mikrotikAdapter) ListPPPoEProfiles(a any) error {
	c := a.(*gin.Context)

	nas, password, err := h.getNasFromParam(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	results, err := h.mikrotikPort().ListPPPoEProfiles(nas.Host, nas.RestPort, nas.Username, password, nas.UseSSL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: results})
	return nil
}

func (h *mikrotikAdapter) CreatePPPoEProfile(a any) error {
	c := a.(*gin.Context)

	nas, password, err := h.getNasFromParam(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	var payload model.MikrotikProfileInput
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return nil
	}

	err = h.mikrotikPort().CreatePPPoEProfile(nas.Host, nas.RestPort, nas.Username, password, nas.UseSSL, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusCreated, model.Response{Success: true})
	return nil
}

// Hotspot Users

func (h *mikrotikAdapter) ListHotspotUsers(a any) error {
	c := a.(*gin.Context)

	nas, password, err := h.getNasFromParam(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	results, err := h.mikrotikPort().ListHotspotUsers(nas.Host, nas.RestPort, nas.Username, password, nas.UseSSL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: results})
	return nil
}

func (h *mikrotikAdapter) CreateHotspotUser(a any) error {
	c := a.(*gin.Context)

	nas, password, err := h.getNasFromParam(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	var payload model.MikrotikHotspotUserInput
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return nil
	}

	err = h.mikrotikPort().CreateHotspotUser(nas.Host, nas.RestPort, nas.Username, password, nas.UseSSL, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusCreated, model.Response{Success: true})
	return nil
}

func (h *mikrotikAdapter) UpdateHotspotUser(a any) error {
	c := a.(*gin.Context)

	nas, password, err := h.getNasFromParam(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	uid := c.Param("uid")

	var payload model.MikrotikHotspotUserInput
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return nil
	}

	err = h.mikrotikPort().UpdateHotspotUser(nas.Host, nas.RestPort, nas.Username, password, nas.UseSSL, uid, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true})
	return nil
}

func (h *mikrotikAdapter) DeleteHotspotUser(a any) error {
	c := a.(*gin.Context)

	nas, password, err := h.getNasFromParam(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	uid := c.Param("uid")

	err = h.mikrotikPort().DeleteHotspotUser(nas.Host, nas.RestPort, nas.Username, password, nas.UseSSL, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true})
	return nil
}

// Simple Queues

func (h *mikrotikAdapter) ListSimpleQueues(a any) error {
	c := a.(*gin.Context)

	nas, password, err := h.getNasFromParam(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	results, err := h.mikrotikPort().ListSimpleQueues(nas.Host, nas.RestPort, nas.Username, password, nas.UseSSL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: results})
	return nil
}

func (h *mikrotikAdapter) CreateSimpleQueue(a any) error {
	c := a.(*gin.Context)

	nas, password, err := h.getNasFromParam(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	var payload model.MikrotikSimpleQueueInput
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return nil
	}

	err = h.mikrotikPort().CreateSimpleQueue(nas.Host, nas.RestPort, nas.Username, password, nas.UseSSL, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusCreated, model.Response{Success: true})
	return nil
}

func (h *mikrotikAdapter) UpdateSimpleQueue(a any) error {
	c := a.(*gin.Context)

	nas, password, err := h.getNasFromParam(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	qid := c.Param("qid")

	var payload model.MikrotikSimpleQueueInput
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return nil
	}

	err = h.mikrotikPort().UpdateSimpleQueue(nas.Host, nas.RestPort, nas.Username, password, nas.UseSSL, qid, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true})
	return nil
}

func (h *mikrotikAdapter) DeleteSimpleQueue(a any) error {
	c := a.(*gin.Context)

	nas, password, err := h.getNasFromParam(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	qid := c.Param("qid")

	err = h.mikrotikPort().DeleteSimpleQueue(nas.Host, nas.RestPort, nas.Username, password, nas.UseSSL, qid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true})
	return nil
}

// Monitoring

func (h *mikrotikAdapter) GetActiveConnections(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_mikrotik_connections")

	nasID := c.Param("id")

	results, err := h.domain.Mikrotik().GetActiveConnections(ctx, nasID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: results})
	return nil
}

func (h *mikrotikAdapter) GetInterfaceTraffic(a any) error {
	c := a.(*gin.Context)

	nas, password, err := h.getNasFromParam(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	interfaceName := c.Param("interface")

	result, err := h.mikrotikPort().GetInterfaceTraffic(nas.Host, nas.ApiPort, nas.Username, password, nas.UseSSL, interfaceName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: result})
	return nil
}
