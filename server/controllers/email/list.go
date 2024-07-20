package email

import (
	"encoding/json"
	"github.com/Jinnrry/pmail/dto"
	"github.com/Jinnrry/pmail/dto/response"
	"github.com/Jinnrry/pmail/services/list"
	"github.com/Jinnrry/pmail/utils/context"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"io"
	"math"
	"net/http"
)

type emailListResponse struct {
	CurrentPage int         `json:"current_page"`
	TotalPage   int         `json:"total_page"`
	List        []*emilItem `json:"list"`
}

type emilItem struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Desc      string `json:"desc"`
	Datetime  string `json:"datetime"`
	IsRead    bool   `json:"is_read"`
	Sender    User   `json:"sender"`
	To        []User `json:"to"`
	Dangerous bool   `json:"dangerous"`
	Error     string `json:"error"`
}

type User struct {
	Name         string `json:"Name"`
	EmailAddress string `json:"EmailAddress"`
}

type emailRequest struct {
	Keyword     string `json:"keyword"`
	Tag         string `json:"tag"`
	CurrentPage int    `json:"current_page"`
	PageSize    int    `json:"page_size"`
}

func EmailList(ctx *context.Context, w http.ResponseWriter, req *http.Request) {
	var lst []*emilItem
	reqBytes, err := io.ReadAll(req.Body)
	if err != nil {
		log.WithContext(ctx).Errorf("%+v", err)
	}
	var retData emailRequest
	err = json.Unmarshal(reqBytes, &retData)
	if err != nil {
		log.WithContext(ctx).Errorf("%+v", err)
	}

	offset := 0
	if retData.CurrentPage >= 1 {
		offset = (retData.CurrentPage - 1) * retData.PageSize
	}

	if retData.PageSize == 0 {
		retData.PageSize = 15
	}

	var tagInfo dto.SearchTag = dto.SearchTag{
		Type:    -1,
		Status:  -1,
		GroupId: -1,
	}
	_ = json.Unmarshal([]byte(retData.Tag), &tagInfo)

	emailList, total := list.GetEmailList(ctx, tagInfo, retData.Keyword, false, offset, retData.PageSize)

	for _, email := range emailList {
		var sender User
		_ = json.Unmarshal([]byte(email.Sender), &sender)

		var tos []User
		_ = json.Unmarshal([]byte(email.To), &tos)

		lst = append(lst, &emilItem{
			ID:        email.Id,
			Title:     email.Subject,
			Desc:      email.Text.String,
			Datetime:  email.SendDate.Format("2006-01-02 15:04:05"),
			IsRead:    email.IsRead == 1,
			Sender:    sender,
			To:        tos,
			Dangerous: email.SPFCheck == 0 && email.DKIMCheck == 0,
			Error:     email.Error.String,
		})
	}

	ret := emailListResponse{
		CurrentPage: retData.CurrentPage,
		TotalPage:   cast.ToInt(math.Ceil(cast.ToFloat64(total) / cast.ToFloat64(retData.PageSize))),
		List:        lst,
	}
	response.NewSuccessResponse(ret).FPrint(w)
}
