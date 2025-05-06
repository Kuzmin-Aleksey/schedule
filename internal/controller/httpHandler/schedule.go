package httpHandler

import (
	"encoding/json"
	"net/http"
	"schedule/internal/controller/httpHandler/models"
	"schedule/internal/usecase/schedule"
	"strconv"
	"time"
)

// @Summary      Create schedule
// @Description  Создаёт новое расписание
// @Tags         schedule
// @Accept       json
// @Produce      json
// @Param        input body schedule.CreateScheduleDTO true "schedule info"
// @Success      200 {object} schedule.CreateScheduleResponseDTO
// @Failure      400  {object}  errorResponse
// @Failure      500  {object}  errorResponse
// @Router       /schedule [post]
func (h *Handler) createSchedule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req := new(models.CreateScheduleRequest)
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		h.writeAndLogErr(ctx, w, err, http.StatusBadRequest)
		return
	}

	duration, err := time.ParseDuration(req.Period)
	if err != nil {
		h.writeAndLogErr(ctx, w, err, http.StatusBadRequest)
		return
	}

	reqDto := &schedule.CreateScheduleDTO{
		UserId:   int64(req.UserId),
		Name:     req.Name,
		Duration: uint(req.Duration),
		Period:   duration,
	}

	if err := reqDto.Validate(); err != nil {
		h.writeAndLogErr(ctx, w, err, http.StatusBadRequest)
		return
	}

	respDto, err := h.schedule.Create(r.Context(), reqDto)
	if err != nil {
		h.writeAndLogErr(ctx, w, err, http.StatusInternalServerError)
		return
	}

	resp := &models.CreateScheduleResponse{
		Id: respDto.Id,
	}

	h.writeJson(ctx, w, resp, http.StatusOK)
}

// @Summary      Get user schedules
// @Description  Возвращает список идентификаторов существующих расписаний для указанного пользователя
// @Tags         schedule
// @Accept       json
// @Produce      json
// @Param        user_id query integer  true "user id"
// @Param 		 TZ header string false "timezone" default(+00:00)
// @Success      200  {array}   integer
// @Failure      400  {object}  errorResponse
// @Failure      500  {object}  errorResponse
// @Router       /schedules [get]
func (h *Handler) getUserSchedules(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userId, err := strconv.ParseInt(r.FormValue("user_id"), 10, 64)
	if err != nil {
		h.writeAndLogErr(ctx, w, err, http.StatusBadRequest)
		return
	}

	resp, err := h.schedule.GetByUser(ctx, userId)
	if err != nil {
		h.writeAndLogErr(ctx, w, err, http.StatusInternalServerError)
		return
	}

	h.writeJson(ctx, w, resp, http.StatusOK)
}

// @Summary      Get schedule
// @Description  Возвращает данные о выбранном расписании с рассчитанным графиком приёмов на день
// @Tags         schedule
// @Accept       json
// @Produce      json
// @Param        user_id query integer  true "user id"
// @Param        schedule_id query integer  true "schedule id"
// @Param 		 TZ header string false "timezone" default(+00:00)
// @Success      200  {object}  schedule.ScheduleResponseDTO
// @Failure      400  {object}  errorResponse
// @Failure      500  {object}  errorResponse
// @Router       /schedule [get]
func (h *Handler) getSchedule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userId, err := strconv.ParseInt(r.FormValue("user_id"), 10, 64)
	if err != nil {
		h.writeAndLogErr(ctx, w, err, http.StatusBadRequest)
		return
	}
	scheduleId, err := strconv.Atoi(r.FormValue("schedule_id"))
	if err != nil {
		h.writeAndLogErr(ctx, w, err, http.StatusBadRequest)
		return
	}

	respDto, err := h.schedule.GetTimetable(ctx, userId, scheduleId)
	if err != nil {
		h.writeAndLogErr(ctx, w, err, http.StatusInternalServerError)
		return
	}

	var endAt string
	if respDto.EndAt != nil {
		endAt = respDto.EndAt.String()
	}

	timeTable := make([]string, len(respDto.Timetable))
	for i, t := range respDto.Timetable {
		timeTable[i] = t.String()
	}

	resp := &models.ScheduleResponse{
		Id:        respDto.Id,
		EndAt:     endAt,
		Name:      respDto.Name,
		Period:    respDto.Period.String(),
		Timetable: timeTable,
	}

	h.writeJson(ctx, w, resp, http.StatusOK)
}

// @Summary      Get next takings
// @Description  Возвращает данные о расписаниях на ближайший период
// @Tags         schedule
// @Accept       json
// @Produce      json
// @Param        user_id query integer  true "user id"
// @Param 		 TZ header string false "timezone" default(+00:00)
// @Success      200  {array}   schedule.NextTakingResponseDTO
// @Failure      400  {object}  errorResponse
// @Failure      500  {object}  errorResponse
// @Router       /next_taking [get]
func (h *Handler) scheduleGetNextTakings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userId, err := strconv.ParseInt(r.FormValue("user_id"), 10, 64)
	if err != nil {
		h.writeAndLogErr(ctx, w, err, http.StatusBadRequest)
		return
	}

	respDto, err := h.schedule.GetNextTakings(ctx, userId)
	if err != nil {
		h.writeAndLogErr(ctx, w, err, http.StatusInternalServerError)
		return
	}

	resp := make([]*models.NextTakingResponse, len(respDto))

	for i, t := range respDto {
		var endAt string
		if t.EndAt != nil {
			endAt = t.EndAt.String()
		}

		resp[i] = &models.NextTakingResponse{
			Id:         t.Id,
			EndAt:      endAt,
			Name:       t.Name,
			NextTaking: t.NextTaking.String(),
			Period:     t.Period.String(),
		}
	}

	h.writeJson(ctx, w, resp, http.StatusOK)
}
