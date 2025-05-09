package httpserver

type Server struct {
	ScheduleServer
}

func NewServer(scheduleServer ScheduleServer) *Server {
	var h = &Server{
		ScheduleServer: scheduleServer,
	}

	return h
}
