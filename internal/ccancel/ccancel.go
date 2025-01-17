/**
 * Copyright (c) 2023 Peking University and Peking University
 * Changsha Institute for Computing and Digital Economy
 *
 * CraneSched is licensed under Mulan PSL v2.
 * You can use this software according to the terms and conditions of
 * the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *          http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS,
 * WITHOUT WARRANTIES OF ANY KIND,
 * EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
 * MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
 * See the Mulan PSL v2 for more details.
 */

package ccancel

import (
	"CraneFrontEnd/generated/protos"
	"CraneFrontEnd/internal/util"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var (
	stub protos.CraneCtldClient
)

func CancelTask(args []string) util.CraneCmdError {
	req := &protos.CancelTaskRequest{
		OperatorUid: uint32(os.Getuid()),

		FilterPartition: FlagPartition,
		FilterAccount:   FlagAccount,
		FilterTaskName:  FlagJobName,
		FilterState:     protos.TaskStatus_Invalid,
		FilterUsername:  FlagUserName,
	}

	if len(args) > 0 {
		taskIdStrSplit := strings.Split(args[0], ",")
		var taskIds []uint32
		for i := 0; i < len(taskIdStrSplit); i++ {
			taskId64, err := strconv.ParseUint(taskIdStrSplit[i], 10, 32)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Invalid job Id: "+taskIdStrSplit[i])
				return util.ErrorCmdArg
			}
			taskIds = append(taskIds, uint32(taskId64))
		}
		req.FilterTaskIds = taskIds
	}

	if FlagState != "" {
		FlagState = strings.ToLower(FlagState)
		if FlagState == "pd" || FlagState == "pending" {
			req.FilterState = protos.TaskStatus_Pending
		} else if FlagState == "r" || FlagState == "running" {
			req.FilterState = protos.TaskStatus_Running
		} else {
			_, _ = fmt.Fprintf(os.Stderr, "Invalid FlagState, Valid job states are PENDING, RUNNING.")
			return util.ErrorCmdArg
		}
	}

	req.FilterNodes = FlagNodes

	reply, err := stub.CancelTask(context.Background(), req)
	if err != nil {
		util.GrpcErrorPrintf(err, "Failed to cancel tasks")
		os.Exit(util.ErrorNetwork)
	}

	if len(reply.CancelledTasks) > 0 {
		cancelledTasksStr := strconv.FormatUint(uint64(reply.CancelledTasks[0]), 10)
		for i := 1; i < len(reply.CancelledTasks); i++ {
			cancelledTasksStr += ","
			cancelledTasksStr += strconv.FormatUint(uint64(reply.CancelledTasks[i]), 10)
		}
		fmt.Printf("Job %s cancelled successfully.\n", cancelledTasksStr)
	}

	if len(reply.NotCancelledTasks) > 0 {
		for i := 0; i < len(reply.NotCancelledTasks); i++ {
			_, _ = fmt.Fprintf(os.Stderr, "Failed to cancel job: %d. Reason: %s.\n",
				reply.NotCancelledTasks[i], reply.NotCancelledReasons[i])
		}
		os.Exit(util.ErrorBackend)
	}
	return util.ErrorSuccess
}
