package main

type NumericMessage struct {
	number   int
	template string
}

var (
	ERR_NOSUCHNICK		 = NumericMessage{401, "%s :No such nick or channel name"}
	ERR_UNKNOWNCOMMAND   = NumericMessage{421, "%s :Unknown command"}
	ERR_NONICKNAMEGIVEN  = NumericMessage{431, ":No nickname given"}
	ERR_ERRONEUSNICKNAME = NumericMessage{432, "%s :Erroneous nickname"}
	ERR_NICKNAMEINUSE    = NumericMessage{433, "%s :Nickname is already in use"}
	ERR_NOTREGISTERED    = NumericMessage{451, "%s :Connection not registered"}
)
