package com.aurora.util;

import lombok.extern.log4j.Log4j2;

import java.util.ArrayList;
import java.util.List;

@Log4j2
public class BeanCopyUtil {

    public static <T> T copyObject(Object source, Class<T> target) {
        T temp = null;
        try {
            temp = target.getDeclaredConstructor().newInstance();
            if (null != source) {
                org.springframework.beans.BeanUtils.copyProperties(source, temp);
            }
        } catch (Exception e) {
            log.error("Failed to copy object to {}", target.getName(), e);
        }
        return temp;
    }

    public static <T, S> List<T> copyList(List<S> source, Class<T> target) {
        List<T> list = new ArrayList<>();
        if (null != source && source.size() > 0) {
            for (Object obj : source) {
                list.add(BeanCopyUtil.copyObject(obj, target));
            }
        }
        return list;
    }

}