package com.aurora.service.impl;

import com.aurora.model.dto.UserDetailsDTO;
import com.aurora.service.RedisService;
import com.aurora.service.TokenService;
import com.fasterxml.jackson.databind.ObjectMapper;
import io.jsonwebtoken.Claims;
import io.jsonwebtoken.Jwts;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;
import org.springframework.util.StringUtils;

import javax.crypto.SecretKey;
import javax.crypto.spec.SecretKeySpec;
import jakarta.servlet.http.HttpServletRequest;
import java.time.Duration;
import java.time.LocalDateTime;
import java.util.Base64;
import java.util.Map;
import java.util.Optional;
import java.util.UUID;

import static com.aurora.constant.AuthConstant.*;
import static com.aurora.constant.RedisConstant.LOGIN_USER;


@Service
public class TokenServiceImpl implements TokenService {

    @Value("${jwt.secret}")
    private String secret;

    @Autowired
    private RedisService redisService;

    private final ObjectMapper objectMapper = new ObjectMapper();

    @Override
    public String createToken(UserDetailsDTO userDetailsDTO) {
        refreshToken(userDetailsDTO);
        String userId = userDetailsDTO.getId().toString();
        return createToken(userId);
    }

    @Override
    public String createToken(String subject) {
        SecretKey secretKey = generalKey();
        return Jwts.builder()
                .id(getUuid())
                .subject(subject)
                .issuer("huaweimian")
                .signWith(secretKey)
                .compact();
    }

    @Override
    public void refreshToken(UserDetailsDTO userDetailsDTO) {
        LocalDateTime currentTime = LocalDateTime.now();
        userDetailsDTO.setExpireTime(currentTime.plusSeconds(EXPIRE_TIME));
        String userId = userDetailsDTO.getId().toString();
        redisService.hSet(LOGIN_USER, userId, userDetailsDTO, EXPIRE_TIME);
    }

    @Override
    public void renewToken(UserDetailsDTO userDetailsDTO) {
        LocalDateTime expireTime = userDetailsDTO.getExpireTime();
        LocalDateTime currentTime = LocalDateTime.now();
        if (Duration.between(currentTime, expireTime).toMinutes() <= TWENTY_MINUTES) {
            refreshToken(userDetailsDTO);
        }
    }

    @Override
    public Claims parseToken(String token) {
        SecretKey secretKey = generalKey();
        return Jwts.parser()
                .verifyWith(secretKey)
                .build()
                .parseSignedClaims(token)
                .getPayload();
    }

    @Override
    public UserDetailsDTO getUserDetailDTO(HttpServletRequest request) {
        String token = Optional.ofNullable(request.getHeader(TOKEN_HEADER)).orElse("").replaceFirst(TOKEN_PREFIX, "");
        if (StringUtils.hasText(token) && !token.equals("null")) {
            Claims claims = parseToken(token);
            String userId = claims.getSubject();
            Object obj = redisService.hGet(LOGIN_USER, userId);
            if (obj == null) {
                return null;
            }
            // 如果已经是 UserDetailsDTO 类型，直接返回
            if (obj instanceof UserDetailsDTO) {
                return (UserDetailsDTO) obj;
            }
            // 如果是 Map 类型（从 Redis 反序列化得到），转换为 UserDetailsDTO
            if (obj instanceof Map) {
                try {
                    return objectMapper.convertValue(obj, UserDetailsDTO.class);
                } catch (Exception e) {
                    e.printStackTrace();
                    return null;
                }
            }
            return null;
        }
        return null;
    }

    @Override
    public void delLoginUser(Integer userId) {
        redisService.hDel(LOGIN_USER, String.valueOf(userId));
    }

    public String getUuid() {
        return UUID.randomUUID().toString().replaceAll("-", "");
    }

    public SecretKey generalKey() {
        byte[] encodedKey = Base64.getDecoder().decode(secret);
        return new SecretKeySpec(encodedKey, "HmacSHA256");
    }

}
