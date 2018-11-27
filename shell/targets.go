package shell

import (
	"errors"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var (
	regTargetSingle    = regexp.MustCompile(`^[0-9]+$`)
	regTargetMulti     = regexp.MustCompile(`^[0-9\,]+$`)
	regTargetRange     = regexp.MustCompile(`^[0-9]+\-[0-9]+$`)
	regTargetRangeUp   = regexp.MustCompile(`^[0-9]+\-$`)
	regTargetRangeDown = regexp.MustCompile(`^\-[0-9]+$`)
)

func (s *Shell) getTargetsFromSet(targetSet string) []*socket {
	targets := []*socket{}

	if targetSet == "" {
		return targets
	}

	targetIDs := []int{}

	s.socketMu.Lock()
	defer s.socketMu.Unlock()

	switch {
	case targetSet == "*":
		for id, _ := range s.openSockets {
			i, _ := strconv.Atoi(id)
			targetIDs = append(targetIDs, i)
		}
	case regTargetSingle.MatchString(targetSet):
		if _, ok := s.openSockets[targetSet]; ok {
			i, _ := strconv.Atoi(targetSet)
			targetIDs = append(targetIDs, i)
		}
	case regTargetMulti.MatchString(targetSet):
		tSet := strings.Split(targetSet, ",")
		for _, id := range tSet {
			if _, ok := s.openSockets[id]; ok {
				i, _ := strconv.Atoi(id)
				targetIDs = append(targetIDs, i)
			}
		}
	case regTargetRange.MatchString(targetSet):
		bounds := strings.Split(targetSet, "-")
		bLow, _ := strconv.Atoi(bounds[0])
		bHigh, _ := strconv.Atoi(bounds[1])

		for id, _ := range s.openSockets {
			i, _ := strconv.Atoi(id)
			if i >= bLow && i <= bHigh {
				targetIDs = append(targetIDs, i)
			}
		}
	case regTargetRangeDown.MatchString(targetSet):
		bHigh, _ := strconv.Atoi(strings.Trim(targetSet, "-"))

		for id, _ := range s.openSockets {
			i, _ := strconv.Atoi(id)
			if i <= bHigh {
				targetIDs = append(targetIDs, i)
			}
		}
	case regTargetRangeUp.MatchString(targetSet):
		bLow, _ := strconv.Atoi(strings.Trim(targetSet, "-"))

		for id, _ := range s.openSockets {
			i, _ := strconv.Atoi(id)
			if i >= bLow {
				targetIDs = append(targetIDs, i)
			}
		}
	}

	sort.Ints(targetIDs)

	for _, i := range targetIDs {
		id := strconv.Itoa(i)
		if sock, ok := s.openSockets[id]; ok {
			targets = append(targets, sock)
		} else {
			//sanity check
			panic("socket " + id + " does not exist, this shouldn't happen")
		}
	}

	return targets
}

func (s *Shell) getTargets() []*socket {
	targetSet := s.cctx.Get("targets")
	return s.getTargetsFromSet(targetSet)
}

func validTargetSet(targetSet string) bool {
	switch {
	case targetSet == "*":
		return true
	case regTargetSingle.MatchString(targetSet):
		return true
	case regTargetMulti.MatchString(targetSet):
		return true
	case regTargetRange.MatchString(targetSet):
		return true
	case regTargetRangeDown.MatchString(targetSet):
		return true
	case regTargetRangeUp.MatchString(targetSet):
		return true
	}

	return false
}

func (s *Shell) sendToTargets(data []byte) (int, []*socket) {
	targets := s.getTargets()
	if len(targets) == 0 {
		s.Err(errors.New("no matching targets to send to"))
		return 0, targets
	}

	count := 0
	for _, t := range targets {
		s.consoleWriteStrPrefix("sending to websocket: " + t.id() + "\n")
		err := t.write(data)
		if err != nil {
			s.Err(err)
			continue
		}
		count++
	}

	return count, targets
}

func (s *Shell) sendToTargetsAck(data []byte, ackfn ackFunc) (int, []*socket) {
	targets := s.getTargets()
	if len(targets) == 0 {
		s.Err(errors.New("no matching targets to send to"))
		return 0, targets
	}

	count := 0
	for _, t := range targets {
		s.consoleWriteStrPrefix("sending to websocket: " + t.id() + "\n")
		err := t.writeAck(data, ackfn)
		if err != nil {
			s.Err(err)
			continue
		}
		count++
	}

	return count, targets
}
