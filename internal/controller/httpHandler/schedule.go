package httpHandler

import (
	"encoding/json"
	"net/http"
	"schedule/internal/usecase/schedule"
	"strconv"
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
	dto := new(schedule.CreateScheduleDTO)
	if err := json.NewDecoder(r.Body).Decode(dto); err != nil {
		h.writeAndLogErr(w, err, http.StatusBadRequest)
		return
	}
	if err := dto.Validate(); err != nil {
		h.writeAndLogErr(w, err, http.StatusBadRequest)
		return
	}

	resp, err := h.schedule.Create(r.Context(), dto)
	if err != nil {
		h.writeAndLogErr(w, err, http.StatusInternalServerError)
		return
	}

	h.writeJson(w, resp, http.StatusOK)
}

// @Summary      Get user schedules
// @Description  Возвращает список идентификаторов существующих расписаний для указанного пользователя
// @Tags         schedule
// @Accept       json
// @Produce      json
// @Param        user_id query integer  true "user id"
// @Success      200  {array}   integer
// @Failure      400  {object}  errorResponse
// @Failure      500  {object}  errorResponse
// @Router       /schedules [get]
func (h *Handler) getUserSchedules(w http.ResponseWriter, r *http.Request) {
	userId, err := strconv.ParseInt(r.FormValue("user_id"), 10, 64)
	if err != nil {
		h.writeAndLogErr(w, err, http.StatusBadRequest)
		return
	}

	resp, err := h.schedule.GetByUser(r.Context(), userId)
	if err != nil {
		h.writeAndLogErr(w, err, http.StatusInternalServerError)
		return
	}

	h.writeJson(w, resp, http.StatusOK)
}

// @Summary      Get schedule
// @Description  Возвращает данные о выбранном расписании с рассчитанным графиком приёмов на день
// @Tags         schedule
// @Accept       json
// @Produce      json
// @Param        user_id query integer  true "user id"
// @Param        schedule_id query integer  true "schedule id"
// @Success      200  {object}  schedule.ScheduleResponseDTO
// @Failure      400  {object}  errorResponse
// @Failure      500  {object}  errorResponse
// @Router       /schedule [get]
func (h *Handler) getSchedule(w http.ResponseWriter, r *http.Request) {
	userId, err := strconv.ParseInt(r.FormValue("user_id"), 10, 64)
	if err != nil {
		h.writeAndLogErr(w, err, http.StatusBadRequest)
		return
	}
	scheduleId, err := strconv.Atoi(r.FormValue("schedule_id"))
	if err != nil {
		h.writeAndLogErr(w, err, http.StatusBadRequest)
		return
	}

	resp, err := h.schedule.GetTimetable(r.Context(), userId, scheduleId)
	if err != nil {
		h.writeAndLogErr(w, err, http.StatusInternalServerError)
		return
	}

	h.writeJson(w, resp, http.StatusOK)
}

// @Summary      Get next takings
// @Description  Возвращает данные о расписаниях на ближайший период
// @Tags         schedule
// @Accept       json
// @Produce      json
// @Param        user_id query integer  true "user id"
// @Success      200  {array}   schedule.NextTakingResponseDTO
// @Failure      400  {object}  errorResponse
// @Failure      500  {object}  errorResponse
// @Router       /next_taking [get]
func (h *Handler) scheduleGetNextTakings(w http.ResponseWriter, r *http.Request) {
	userId, err := strconv.ParseInt(r.FormValue("user_id"), 10, 64)
	if err != nil {
		h.writeAndLogErr(w, err, http.StatusBadRequest)
		return
	}

	resp, err := h.schedule.GetNextTakings(r.Context(), userId)
	if err != nil {
		h.writeAndLogErr(w, err, http.StatusInternalServerError)
		return
	}

	h.writeJson(w, resp, http.StatusOK)
}
